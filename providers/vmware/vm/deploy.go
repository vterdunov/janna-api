package vm

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/vmware/govmomi/nfc"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vterdunov/janna-api/config"
)

type deployment struct {
	Client *vim25.Client
	Finder *find.Finder
	ovfx
}

type ovfx struct {
	Name         string
	Datacenter   *object.Datacenter
	Datastore    *object.Datastore
	ResourcePool *object.ResourcePool
	Folder       *object.Folder
	Cluster      *object.ClusterComputeResource
	Host         *object.HostSystem
}

func (o *deployment) ChooseDatacenter(ctx context.Context, dcName string) error {
	dc, err := o.Finder.DatacenterOrDefault(ctx, dcName)
	if err != nil {
		return err
	}
	o.Finder.SetDatacenter(dc)
	o.Datacenter = dc
	return nil
}

func (o *deployment) ChooseDatastore(ctx context.Context, dsName string) error {
	// TODO: try to use DatastoreCLuster instead of Datastore
	//   user can choose that want to use
	ds, err := o.Finder.DatastoreOrDefault(ctx, dsName)
	if err != nil {
		return err
	}
	o.Datastore = ds
	return nil
}

func (o *deployment) ChooseResourcePool(ctx context.Context, rpName string) error {
	rp, err := o.Finder.ResourcePoolOrDefault(ctx, rpName)
	if err != nil {
		return err
	}
	o.ResourcePool = rp
	return nil
}

func (o *deployment) ChooseFolder(ctx context.Context, fName string) error {
	folder, err := o.Finder.FolderOrDefault(ctx, fName)
	if err != nil {
		return err
	}
	o.Folder = folder
	return nil
}

func (o *deployment) ChooseHost(ctx context.Context, hName string) error {
	// Host param is optional. If we use 'nil', then vCenter will choose a host
	// If you need a specify a cluster then specify a Resource Pool param.
	if hName == "" {
		o.Host = nil
		return nil
	}

	host, err := o.Finder.HostSystem(ctx, hName)
	if err != nil {
		return err
	}
	o.Host = host
	return nil
}

func (o *deployment) NetworkMap(e *ovf.Envelope) (p []types.OvfNetworkMapping) {
	ctx := context.TODO()
	networks := map[string]string{}

	if e.Network != nil {
		for _, net := range e.Network.Networks {
			networks[net.Name] = net.Name
		}
	}

	// TODO: rewrite networks from params

	for src, dst := range networks {
		if net, err := o.Finder.Network(ctx, dst); err == nil {
			p = append(p, types.OvfNetworkMapping{
				Name:    src,
				Network: net.Reference(),
			})
		}
	}
	return
}

func (o *deployment) Upload(ctx context.Context, lease *nfc.Lease, item nfc.FileItem) error {
	file := item.Path

	f, err := os.Open(file)
	if err != nil {
		return err
	}

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	defer f.Close()

	outputStr := fmt.Sprintf("Uploading %s... ", path.Base(file))
	pl := newProgressLogger(outputStr)
	defer pl.Wait()

	opts := soap.Upload{
		ContentLength: stat.Size(),
		Progress:      pl,
	}

	return lease.Upload(ctx, item, f, opts)
}

func (o *deployment) Import(ctx context.Context, OVAURL string) (*types.ManagedObjectReference, error) {

	url, err := url.Parse(OVAURL)
	if err != nil {
		return nil, err
	}

	rovf, _, err := o.Client.Download(url, &soap.DefaultDownload)
	if err != nil {
		return nil, err
	}

	td, err := ioutil.TempDir("", "janna-")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(td)

	if untarErr := Untar(td, rovf); untarErr != nil {
		return nil, untarErr
	}

	ovfName, err := checkExt(".ovf", td)
	if err != nil {
		return nil, err
	}

	ovfPath := td + "/" + ovfName
	fmt.Println(ovfPath)
	rova, err := os.Open(ovfPath)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(rova)
	if err != nil {
		return nil, err
	}

	defer rovf.Close()

	e, err := readEnvelope(b)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read Envelope")
	}

	name := "Virtual Appliance"
	if e.VirtualSystem != nil {
		name = e.VirtualSystem.ID
		if e.VirtualSystem.Name != nil {
			name = *e.VirtualSystem.Name
		}
	}

	// Override name from params if specified
	if o.Name != "" {
		name = o.Name
	}

	cisp := types.OvfCreateImportSpecParams{
		// See https://github.com/vmware/govmomi/blob/v0.16.0/vim25/types/enum.go#L3381-L3395
		// VMWare can not support some of those disk format types
		// "preallocated", "thin", "seSparse", "rdm", "rdmp",
		// "raw", "delta", "sparse2Gb", "thick2Gb", "eagerZeroedThick",
		// "sparseMonolithic", "flatMonolithic", "thick"
		DiskProvisioning: "thin",
		EntityName:       name,
		NetworkMapping:   o.NetworkMap(e),
	}

	m := ovf.NewManager(o.Client)
	rp := o.ResourcePool
	ds := o.Datastore
	spec, err := m.CreateImportSpec(ctx, string(b), rp, ds, cisp)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create VM spec")
	}
	if spec.Error != nil {
		return nil, errors.New(spec.Error[0].LocalizedMessage)
	}

	// TODO: get from params
	anno := "Test annotations"
	if anno != "" {
		switch s := spec.ImportSpec.(type) {
		case *types.VirtualMachineImportSpec:
			s.ConfigSpec.Annotation = anno
		case *types.VirtualAppImportSpec:
			s.VAppConfigSpec.Annotation = anno
		}
	}

	lease, err := rp.ImportVApp(ctx, spec.ImportSpec, o.Folder, o.Host)
	if err != nil {
		return nil, errors.Wrap(err, "Could not import Virtual Appliance")
	}

	info, err := lease.Wait(ctx, spec.FileItem)
	if err != nil {
		return nil, err
	}

	u := lease.StartUpdater(ctx, info)
	defer u.Done()

	for _, item := range info.Items {
		if err = o.Upload(ctx, lease, item); err != nil {
			return nil, errors.Wrap(err, "Could not upload disks to VMWare")
		}
	}
	return &info.Entity, lease.Complete(ctx)
}

// Deploy Virtual Machine to VMWare
func Deploy(ctx context.Context, vmName string, OVAURL string, logger log.Logger, cfg *config.Config, c *vim25.Client, opts ...string) (int, error) {
	// TODO: make up a metod to check deploy progress.
	// Job ID and endpoint with status?
	// keep HTTP connection with client and poll it?
	var jid int

	logger.Log("msg", "Starting deploy VM", "vm", vmName)

	d := newDeployment(c, vmName)
	if err := d.ChooseDatacenter(ctx, cfg.VMWare.DC); err != nil {
		return jid, err
	}

	if err := d.ChooseDatastore(ctx, cfg.VMWare.DS); err != nil {
		return jid, err
	}

	if err := d.ChooseResourcePool(ctx, cfg.VMWare.RP); err != nil {
		return jid, err
	}

	if err := d.ChooseFolder(ctx, cfg.VMWare.Folder); err != nil {
		return jid, err
	}

	if err := d.ChooseHost(ctx, cfg.VMWare.Host); err != nil {
		return jid, err
	}

	// TODO: Download OVF. Download and unpack OVA
	moref, err := d.Import(ctx, OVAURL)
	if err != nil {
		return jid, err
	}

	vm := object.NewVirtualMachine(c, *moref)

	logger.Log("msg", "Powering on...", "vm", vmName)
	if err = powerON(ctx, vm); err != nil {
		return jid, err
	}

	// WaitForIP
	logger.Log("msg", "Waiting for ip", "vm", vmName)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	ip, err := waitForIP(ctx, vm)
	if err != nil {
		logger.Log("err", errors.Wrap(err, "Could not get IP address"), "vm", vmName)
		return jid, err
	}
	logger.Log("msg", "Received IP address", "vm", vmName, "ip", ip)

	return jid, nil
}

func newDeployment(c *vim25.Client, vmName string) *deployment {
	finder := find.NewFinder(c, true)
	return &deployment{
		Client: c,
		Finder: finder,
		ovfx:   ovfx{Name: vmName},
	}
}

func powerON(ctx context.Context, vm *object.VirtualMachine) error {
	task, err := vm.PowerOn(ctx)
	if err != nil {
		return errors.Wrap(err, "Could not power on VM")
	}
	if _, err := task.WaitForResult(ctx, nil); err != nil {
		return errors.Wrap(err, "Failed while powering on task")
	}

	return nil
}

func waitForIP(ctx context.Context, vm *object.VirtualMachine) (string, error) {
	ip, err := vm.WaitForIP(ctx)
	if err != nil {
		return "", err
	}
	return ip, nil
}

// func readOVF(fpath string) ([]byte, error) {
// 	f, err := os.Open(fpath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer f.Close()

// 	r, err := ioutil.ReadAll(f)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return r, nil
// }

func readEnvelope(data []byte) (*ovf.Envelope, error) {
	r := bytes.NewReader(data)

	e, err := ovf.Unmarshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse ovf")
	}

	return e, nil
}

type progressLogger struct {
	prefix string

	wg sync.WaitGroup

	sink chan chan progress.Report
	done chan struct{}
}

func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	called := false

	for stop := false; !stop; {
		select {
		case ch := <-p.sink:
			err = p.loopB(tick, ch)
			stop = true
			called = true
		case <-p.done:
			stop = true
		case <-tick.C:
			line := fmt.Sprintf("\r%s", p.prefix)
			fmt.Println(line)
		}
	}

	if err != nil && err != io.EOF {
		fmt.Println(fmt.Sprintf("\r%sError: %s\n", p.prefix, err))
	} else if called {
		fmt.Println(fmt.Sprintf("\r%sOK\n", p.prefix))
	}
}

// loopA runs after Sink() has been called.
func (p *progressLogger) loopB(tick *time.Ticker, ch <-chan progress.Report) error {
	var r progress.Report
	var ok bool
	var err error

	for ok = true; ok; {
		select {
		case r, ok = <-ch:
			if !ok {
				break
			}
			err = r.Error()
		case <-tick.C:
			line := fmt.Sprintf("\r%s", p.prefix)
			if r != nil {
				line += fmt.Sprintf("(%.0f%%", r.Percentage())
				detail := r.Detail()
				if detail != "" {
					line += fmt.Sprintf(", %s", detail)
				}
				line += ")"
			}
			fmt.Println(line)
		}
	}

	return err
}

func (p *progressLogger) Wait() {
	close(p.done)
	p.wg.Wait()
}

func (p *progressLogger) Sink() chan<- progress.Report {
	ch := make(chan progress.Report)
	p.sink <- ch
	return ch
}

func newProgressLogger(prefix string) *progressLogger {
	p := &progressLogger{
		prefix: prefix,

		sink: make(chan chan progress.Report),
		done: make(chan struct{}),
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}

func Untar(dst string, r io.Reader) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
}

func checkExt(ext string, dir string) (string, error) {
	var file string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ext {
			file = f.Name()
		}
		return nil
	})

	return file, nil
}

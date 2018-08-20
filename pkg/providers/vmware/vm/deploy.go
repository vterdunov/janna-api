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

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vterdunov/janna-api/pkg/config"
	jt "github.com/vterdunov/janna-api/pkg/types"
)

type Deployment struct {
	Client *vim25.Client
	Finder *find.Finder
	ovfx
	logger log.Logger
}

type ovfx struct {
	Name           string
	Datacenter     *object.Datacenter
	Datastore      *object.Datastore
	ResourcePool   *object.ResourcePool
	Folder         *object.Folder
	Cluster        *object.ClusterComputeResource
	Host           *object.HostSystem
	NetworkMapping []Network
}

// Network represent mapping between OVF network and ESXi system network.
// 'OVF-VM-Network-Name' -> 'Yours-ESXi-VM-Network-Name'
type Network struct {
	Name    string
	Network string
}

func (o *Deployment) chooseDatacenter(ctx context.Context, dcName string) error {
	dc, err := o.Finder.DatacenterOrDefault(ctx, dcName)
	if err != nil {
		return err
	}
	o.Finder.SetDatacenter(dc)
	o.Datacenter = dc
	return nil
}

func (o *Deployment) chooseDatastore(ctx context.Context, dsName string) error {
	// TODO: try to use DatastoreCLuster instead of Datastore
	//   user can choose that want to use
	ds, err := o.Finder.DatastoreOrDefault(ctx, dsName)
	if err != nil {
		return err
	}
	o.Datastore = ds
	return nil
}

func (o *Deployment) chooseResourcePool(ctx context.Context, rpName string) error {
	rp, err := o.Finder.ResourcePoolOrDefault(ctx, rpName)
	if err != nil {
		return err
	}
	o.ResourcePool = rp
	return nil
}

func (o *Deployment) chooseFolder(ctx context.Context, fName string) error {
	folder, err := o.Finder.FolderOrDefault(ctx, fName)
	if err != nil {
		return err
	}
	o.Folder = folder
	return nil
}

func (o *Deployment) chooseHost(ctx context.Context, hName string) error {
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

func (o *Deployment) networkMap(ctx context.Context, e *ovf.Envelope) (p []types.OvfNetworkMapping) {
	networks := map[string]string{}

	if e.Network != nil {
		for _, net := range e.Network.Networks {
			o.logger.Log("msg", "found OVF networks mapping", "name", net.Name, "network", net.Name)
			networks[net.Name] = net.Name
		}
	}

	for _, net := range o.NetworkMapping {
		o.logger.Log("msg", "found override networks mapping", "name", net.Name, "network", net.Network)
		networks[net.Name] = net.Network
	}

	for src, dst := range networks {
		if net, err := o.Finder.Network(ctx, dst); err == nil {
			p = append(p, types.OvfNetworkMapping{
				Name:    src,
				Network: net.Reference(),
			})
			o.logger.Log("msg", "networks mapping", "name", src, "network", dst)
		}
	}

	return p
}

func (o *Deployment) Upload(ctx context.Context, lease *nfc.Lease, item nfc.FileItem) error {
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

	outputStr := path.Base(file)
	pl := newProgressLogger(outputStr, o.logger)
	defer pl.Wait()

	opts := soap.Upload{
		ContentLength: stat.Size(),
		Progress:      pl,
	}

	return lease.Upload(ctx, item, f, opts)
}

func (o *Deployment) Import(ctx context.Context, OVAURL string) (*types.ManagedObjectReference, error) {
	// FIXME: logger is nil
	o.logger.Log("test", "test")
	url, err := url.Parse(OVAURL)
	if err != nil {
		return nil, err
	}

	rovf, _, err := o.Client.Download(ctx, url, &soap.DefaultDownload)
	if err != nil {
		o.logger.Log("err", err)
		return nil, err
	}

	td, err := ioutil.TempDir("", "janna-")
	if err != nil {
		o.logger.Log("err", err)
		return nil, err
	}

	defer os.RemoveAll(td)

	if untarErr := untar(td, rovf); untarErr != nil {
		o.logger.Log("err", untarErr)
		return nil, untarErr
	}

	ovfName, err := checkOVFfiles(td)
	if err != nil {
		o.logger.Log("err", err)
		return nil, err
	}

	ovfPath := td + "/" + ovfName
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
		// TODO: get form params
		DiskProvisioning: "thin",
		EntityName:       name,
		NetworkMapping:   o.networkMap(ctx, e),
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

	os.Chdir(td)
	for _, item := range info.Items {
		if err = o.Upload(ctx, lease, item); err != nil {
			return nil, errors.Wrap(err, "Could not upload disks to VMWare")
		}
	}
	return &info.Entity, lease.Complete(ctx)
}

func IsVMExist(ctx context.Context, c *vim25.Client, params *jt.VMDeployParams) (bool, error) {
	f := find.NewFinder(c, false)
	dc, err := f.DatacenterOrDefault(ctx, params.Datacenter)
	if err != nil {
		return false, err
	}
	f.SetDatacenter(dc)

	_, err = f.VirtualMachine(ctx, params.Name)
	switch err.(type) {
	case *find.NotFoundError:
		return false, nil
	default:
		return true, err
	}
}

// NewDeployment create a new deployment object.
// It choose needed resources
func NewDeployment(ctx context.Context, c *vim25.Client, params *jt.VMDeployParams, l log.Logger, cfg *config.Config) (*Deployment, error) { // nolint: unparam
	d := newSimpleDeployment(c, params, l)
	if err := d.chooseDatacenter(ctx, cfg.VMWare.DC); err != nil {
		l.Log("err", errors.Wrap(err, "Could not choose datacenter"))
		return nil, err
	}

	if err := d.chooseDatastore(ctx, cfg.VMWare.DS); err != nil {
		l.Log("err", errors.Wrap(err, "Could not choose datastore"))
		return nil, err
	}

	if err := d.chooseResourcePool(ctx, cfg.VMWare.RP); err != nil {
		l.Log("err", errors.Wrap(err, "Could not choose resource pool"))
		return nil, err
	}

	if err := d.chooseFolder(ctx, cfg.VMWare.Folder); err != nil {
		l.Log("err", errors.Wrap(err, "Could not choose folder"))
		return nil, err
	}

	if err := d.chooseHost(ctx, cfg.VMWare.Host); err != nil {
		l.Log("err", errors.Wrap(err, "Could not choose host"))
		return nil, err
	}

	return d, nil
}

func newSimpleDeployment(c *vim25.Client, deployParams *jt.VMDeployParams, logger log.Logger) *Deployment {
	finder := find.NewFinder(c, true)
	var nms []Network

	if len(deployParams.Networks) != 0 {
		for name, network := range deployParams.Networks {
			nm := Network{
				Name:    name,
				Network: network,
			}
			nms = append(nms, nm)
		}
	}

	ovf := ovfx{
		Name:           deployParams.Name,
		NetworkMapping: nms,
	}

	d := &Deployment{
		Client: c,
		Finder: finder,
		logger: logger,
		ovfx:   ovf,
	}

	return d
}

func PowerON(ctx context.Context, vm *object.VirtualMachine) error {
	task, err := vm.PowerOn(ctx)
	if err != nil {
		return errors.Wrap(err, "Could not power on VM")
	}
	if _, err := task.WaitForResult(ctx, nil); err != nil {
		return errors.Wrap(err, "Failed while powering on task")
	}

	return nil
}

func WaitForIP(ctx context.Context, vm *object.VirtualMachine) (string, error) {
	ip, err := vm.WaitForIP(ctx)
	if err != nil {
		return "", err
	}
	return ip, nil
}

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

	sink   chan chan progress.Report
	done   chan struct{}
	logger log.Logger
}

func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	tick := time.NewTicker(5 * time.Second)
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
		}
	}

	// TODO: change to using logger
	if err != nil && err != io.EOF {
		p.logger.Log("err", errors.Wrap(err, "Error with disks uploading"), "file", p.prefix)
	} else if called {
		p.logger.Log("msg", "uploaded", "file", p.prefix)
	}
}

// loopB runs after Sink() has been called.
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
			if r != nil {
				pc := fmt.Sprintf("%.0f%%", r.Percentage())
				p.logger.Log("msg", "uploading disks", "file", p.prefix, "progress", pc)
			}
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

func newProgressLogger(prefix string, logger log.Logger) *progressLogger {
	p := &progressLogger{
		prefix: prefix,

		sink:   make(chan chan progress.Report),
		done:   make(chan struct{}),
		logger: logger,
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}

func untar(dst string, r io.Reader) error {
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

func checkOVFfiles(dir string) (string, error) {
	var file string
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".ovf" {
			file = f.Name()
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return file, nil
}

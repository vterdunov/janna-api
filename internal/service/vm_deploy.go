package service

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/soap"
	vmware_types "github.com/vmware/govmomi/vim25/types"

	"github.com/vterdunov/janna-api/internal/types"
)

type Deployment struct {
	Client *vim25.Client
	Finder *find.Finder
	logger log.Logger

	ovfx
}

type ovfx struct {
	// Name is the Virtual Machine name
	Name           string
	Datacenter     *object.Datacenter
	Datastore      *object.Datastore
	Folder         *object.Folder
	ResourcePool   *object.ResourcePool
	Host           *object.HostSystem
	NetworkMapping []Network
	Annotation     string
}

// Network defines a mapping from each network inside the OVF
// to a ESXi network. The networks must be presented on the ESXi host.
type Network struct {
	Name    string
	Network string
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func (s *service) VMDeploy(ctx context.Context, params *types.VMDeployParams) (string, error) {
	// TODO: validate incoming params according business rules (https://github.com/asaskevich/govalidator)
	// use Endpoint middleware

	// predeploy checks
	exist, err := isVMExist(ctx, s.Client, params)
	if err != nil {
		return "", err
	}

	if exist {
		return "", fmt.Errorf("Virtual Machine '%s' already exist", params.Name) //nolint: stylecheck,golint
	}

	reqID, ok := ctx.Value(ContextKeyRequestXRequestID).(string)
	if !ok {
		reqID = ""
	}

	l := log.With(s.logger, "request_id", reqID)
	l = log.With(l, "vm", params.Name)

	taskCtx, cancel := context.WithTimeout(context.Background(), s.cfg.TaskTTL)

	t := s.statuses.NewTask()
	t.Str("stage", "start")

	// Start deploy in background
	go func() {
		defer cancel()
		d, err := newDeployment(taskCtx, s.Client, params, l)
		if err != nil {
			err = errors.Wrap(err, "Could not create deployment object")
			l.Log("err", err)
			t.Str(
				"stage", "error",
				"error", err.Error(),
			)
			cancel()
			return
		}

		t.Str("stage", "import")
		moref, err := d.Import(taskCtx, params.OVAURL, params.Annotation)
		if err != nil {
			err = errors.Wrap(err, "Could not import OVA/OVF")
			l.Log("err", err)

			t.Str(
				"stage", "error",
				"error", err.Error(),
			)

			if err, ok := err.(stackTracer); ok {
				for _, f := range err.StackTrace() {
					fmt.Printf("%+s:%d\n", f, f)
				}
			}

			cancel()
			return
		}

		t.Str("stage", "create")
		vmx := object.NewVirtualMachine(s.Client, *moref)

		l.Log("msg", "Powering on...")
		t.Str("message", "Powerig on")
		if err = PowerON(taskCtx, vmx); err != nil {
			err = errors.Wrap(err, "Could not Virtual Machine power on")
			l.Log("err", err)
			t.Str(
				"stage", "error",
				"error", err.Error(),
			)
			cancel()
			return
		}

		t.Str("message", "Waiting for IP addresses")
		ips, err := WaitForIP(taskCtx, vmx)
		if err != nil {
			err = errors.Wrap(err, "error getting IP address")
			l.Log("err", err)
			t.Str(
				"stage", "error",
				"error", err.Error(),
			)
			cancel()
			return
		}

		l.Log("msg", "Successful deploy", "ips", fmt.Sprintf("%v", ips))
		t.Str(
			"stage", "complete",
			"message", "ok",
		).StrArr("ip", ips)

		cancel()
	}()

	return t.ID(), nil
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

func (o *Deployment) chooseDatastore(ctx context.Context, dsType string, names []string) error {
	switch dsType {
	case "cluster":
		if err := o.chooseDatastoreWithCluster(ctx, names); err != nil {
			return err
		}
	case "datastore":
		if err := o.chooseDatastoreWithDatastore(ctx, names); err != nil {
			return err
		}
	default:
		errors.New("could not recognize datastore type. Possible values are 'cluster', 'datastore'")
	}
	return nil
}

func (o *Deployment) chooseDatastoreWithCluster(ctx context.Context, names []string) error {
	pod, err := o.Finder.DatastoreClusterOrDefault(ctx, pickRandom(names))
	if err != nil {
		return err
	}

	drsEnabled, err := isStorageDRSEnabled(ctx, pod)
	if err != nil {
		return err
	}
	if !drsEnabled {
		return errors.New("storage DRS is not enabled on datastore cluster")
	}

	var vmSpec vmware_types.VirtualMachineConfigSpec
	sps := vmware_types.StoragePlacementSpec{
		Type:         string(vmware_types.StoragePlacementSpecPlacementTypeCreate),
		ResourcePool: vmware_types.NewReference(o.ResourcePool.Reference()),
		PodSelectionSpec: vmware_types.StorageDrsPodSelectionSpec{
			StoragePod: vmware_types.NewReference(pod.Reference()),
		},
		Folder:     vmware_types.NewReference(o.Folder.Reference()),
		ConfigSpec: &vmSpec,
	}

	o.logger.Log("msg", "Acquiring Storage DRS recommendations")
	srm := object.NewStorageResourceManager(o.Client)
	placement, err := srm.RecommendDatastores(ctx, sps)
	if err != nil {
		return err
	}

	recs := placement.Recommendations
	if len(recs) < 1 {
		return errors.New("no storage DRS recommendations were found for the requested action")
	}

	spa, ok := recs[0].Action[0].(*vmware_types.StoragePlacementAction)
	if !ok {
		return errors.New("could not get datastore from DRS recomendation")
	}

	ds := spa.Destination
	var mds mo.Datastore
	err = property.DefaultCollector(o.Client).RetrieveOne(ctx, ds, []string{"name"}, &mds)
	if err != nil {
		return err
	}

	datastore := object.NewDatastore(o.Client, ds)

	o.Datastore = datastore
	return nil
}

func isStorageDRSEnabled(ctx context.Context, pod *object.StoragePod) (bool, error) {
	var props mo.StoragePod
	if err := pod.Properties(ctx, pod.Reference(), nil, &props); err != nil {
		return false, err
	}

	if props.PodStorageDrsEntry == nil {
		return false, nil
	}

	return props.PodStorageDrsEntry.StorageDrsConfig.PodConfig.Enabled, nil
}

func (o *Deployment) chooseDatastoreWithDatastore(ctx context.Context, names []string) error {
	ds, err := o.Finder.DatastoreOrDefault(ctx, pickRandom(names))
	if err != nil {
		return err
	}

	o.Datastore = ds
	return nil
}

func pickRandom(slice []string) string {
	rand.Seed(time.Now().Unix())
	return slice[rand.Intn(len(slice))]
}

func (o *Deployment) chooseFolder(ctx context.Context, fName string) error {
	folder, err := o.Finder.FolderOrDefault(ctx, fName)
	if err != nil {
		return err
	}
	o.Folder = folder
	return nil
}

func (o *Deployment) chooseComputerResource(ctx context.Context, resType, path string) error {
	switch resType {
	case "host":
		if err := o.computerResourceWithHost(ctx, path); err != nil {
			return err
		}
	case "cluster":
		if err := o.computerResourceWithCluster(ctx, path); err != nil {
			return err
		}
	case "rp":
		if err := o.computerResourceWithResourcePool(ctx, path); err != nil {
			return err
		}
	default:
		return errors.New("could not recognize computer resource type. Possible types are 'host', 'cluster', 'rp'")
	}

	return nil
}

func (o *Deployment) computerResourceWithHost(ctx context.Context, path string) error {
	host, err := o.Finder.HostSystemOrDefault(ctx, path)
	if err != nil {
		return err
	}

	rp, err := host.ResourcePool(ctx)
	if err != nil {
		return err
	}

	o.Host = host
	o.ResourcePool = rp
	return nil
}

func (o *Deployment) computerResourceWithCluster(ctx context.Context, path string) error {
	cluster, err := o.Finder.ClusterComputeResourceOrDefault(ctx, path)
	if err != nil {
		return err
	}

	rp, err := cluster.ResourcePool(ctx)
	if err != nil {
		return err
	}

	o.ResourcePool = rp

	// vCenter will choose a host
	o.Host = nil
	return nil
}

func (o *Deployment) computerResourceWithResourcePool(ctx context.Context, rpName string) error {
	rp, err := o.Finder.ResourcePoolOrDefault(ctx, rpName)
	if err != nil {
		return err
	}

	o.ResourcePool = rp

	// vCenter will choose a host
	o.Host = nil
	return nil
}

func (o *Deployment) networkMap(ctx context.Context, e *ovf.Envelope) (p []vmware_types.OvfNetworkMapping) {
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
			p = append(p, vmware_types.OvfNetworkMapping{
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

func (o *Deployment) Import(ctx context.Context, ovaURL string, anno string) (*vmware_types.ManagedObjectReference, error) {
	url, err := url.Parse(ovaURL)
	if err != nil {
		return nil, err
	}

	o.logger.Log("msg", "Create temp dir")
	td, err := ioutil.TempDir("", "janna-")
	if err != nil {
		o.logger.Log("err", err)
		return nil, err
	}
	o.logger.Log("msg", "temp dir", "dir", td)
	defer os.RemoveAll(td)
	defer o.logger.Log("msg", "Removed temp dir", "dir", td)

	o.logger.Log("msg", "downloading OVA file", "url", url.String())
	resp, err := grab.Get(td, url.String())
	if err != nil {
		o.logger.Log("err", err)
		return nil, err
	}
	o.logger.Log("OVA file", resp.Filename)

	ova, openErr := os.Open(resp.Filename)
	if openErr != nil {
		return nil, openErr
	}
	defer ova.Close()

	var buferedOVA bytes.Buffer
	tee := io.TeeReader(ova, &buferedOVA)

	hash, hashErr := calculateHash(tee)
	if err != nil {
		o.logger.Log("warn", hashErr)
	}
	o.logger.Log("msg", "downloaded OVA checksumm", "sha256", hash)

	o.logger.Log("msg", "Unpack OVA")
	if untarErr := untar(td, &buferedOVA); untarErr != nil {
		o.logger.Log("err", untarErr)
		return nil, untarErr
	}

	o.logger.Log("msg", "Get OVF path")
	ovfPath, err := findOVF(td)
	if err != nil {
		o.logger.Log("err", err)
		return nil, err
	}
	o.logger.Log("msg", "OVF path", "path", ovfPath)

	o.logger.Log("msg", "Open OVF")
	rawOvf, err := os.Open(ovfPath)
	if err != nil {
		return nil, err
	}

	o.logger.Log("msg", "Read bytes from OVF")
	b, err := ioutil.ReadAll(rawOvf)
	if err != nil {
		return nil, err
	}

	o.logger.Log("msg", "Envelope OVF")
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

	o.logger.Log("msg", "Create Import Spec params")
	cisp := vmware_types.OvfCreateImportSpecParams{
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

	o.logger.Log("msg", "Get OVF manager")
	m := ovf.NewManager(o.Client)
	ovfContent := string(b)
	rp := o.ResourcePool
	ds := o.Datastore

	o.logger.Log("msg", "Create Import Spec")
	spec, err := m.CreateImportSpec(ctx, ovfContent, rp, ds, cisp)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create VM spec")
	}
	if spec.Error != nil {
		o.logger.Log("err", spec.Error[0].LocalizedMessage)
		return nil, errors.New(spec.Error[0].LocalizedMessage)
	}

	if anno != "" {
		switch s := spec.ImportSpec.(type) {
		case *vmware_types.VirtualMachineImportSpec:
			s.ConfigSpec.Annotation = anno
		case *vmware_types.VirtualAppImportSpec:
			s.VAppConfigSpec.Annotation = anno
		}
	}

	o.logger.Log("msg", "Get lease")
	lease, err := rp.ImportVApp(ctx, spec.ImportSpec, o.Folder, o.Host)
	if err != nil {
		err = errors.Wrap(err, "Could not import Virtual Appliance")
		o.logger.Log("err", err)
		return nil, err
	}

	o.logger.Log("msg", "Get lease information")
	info, err := lease.Wait(ctx, spec.FileItem)
	if err != nil {
		err = errors.Wrap(err, "error while waiting lease")
		o.logger.Log("err", err)
		return nil, err
	}

	o.logger.Log("msg", "Get lease updater")
	u := lease.StartUpdater(ctx, info)
	defer u.Done()

	o.logger.Log("msg", "Loop over lease info items")
	for _, item := range info.Items {
		// override disk path to use in cocnurent mode
		// os.Chdir doesn't work preperly
		item.Path = path.Join(td, item.Path)

		o.logger.Log("msg", "Upload disks")
		if err = o.Upload(ctx, lease, item); err != nil {
			return nil, errors.Wrap(err, "Could not upload disks to VMWare")
		}
	}
	o.logger.Log("msg", "End looping over info items")

	return &info.Entity, lease.Complete(ctx)
}

func isVMExist(ctx context.Context, c *vim25.Client, params *types.VMDeployParams) (bool, error) {
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

// newDeployment create a new deployment object.
// It choose needed resources
func newDeployment(ctx context.Context, c *vim25.Client, params *types.VMDeployParams, l log.Logger) (*Deployment, error) { //nolint: unparam
	d := newSimpleDeployment(c, params, l)

	// step 1. choose Datacenter and folder
	if err := d.chooseDatacenter(ctx, params.Datacenter); err != nil {
		err = errors.Wrap(err, "Could not choose datacenter")
		l.Log("err", err)
		return nil, err
	}

	if err := d.chooseFolder(ctx, params.Folder); err != nil {
		err = errors.Wrap(err, "Could not choose folder")
		l.Log("err", err)
		return nil, err
	}

	// step 2. choose computer resource
	resType := params.ComputerResources.Type
	resPath := params.ComputerResources.Path
	if err := d.chooseComputerResource(ctx, resType, resPath); err != nil {
		err = errors.Wrap(err, "Could not choose Computer Resource")
		l.Log("err", err)
		return nil, err
	}

	// step 3. Choose datastore cluster or single datastore
	dsType := params.Datastores.Type
	dsNames := params.Datastores.Names
	if err := d.chooseDatastore(ctx, dsType, dsNames); err != nil {
		err = errors.Wrap(err, "Could not choose datastore")
		l.Log("err", err)
		return nil, err
	}

	return d, nil
}

func newSimpleDeployment(c *vim25.Client, deployParams *types.VMDeployParams, logger log.Logger) *Deployment {
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
		return err
	}
	if _, err := task.WaitForResult(ctx, nil); err != nil {
		return errors.Wrap(err, "Failed while powering on task")
	}

	return nil
}

func WaitForIP(ctx context.Context, vm *object.VirtualMachine) ([]string, error) {
	addresses, err := vm.WaitForNetIP(ctx, true)
	if err != nil {
		return nil, err
	}

	ipAdresses := make([]string, 0)
	for _, ips := range addresses {
		ipAdresses = append(ipAdresses, ips...)
	}

	return ipAdresses, nil
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

	if err != nil && err != io.EOF {
		p.logger.Log("err", errors.Wrap(err, "Error with disks uploading"), "file", p.prefix)
	}

	if called {
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

func findOVF(dir string) (string, error) {
	var ovfPath string
	walcFunc := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".ovf" {
			ovfPath = path
		}

		return nil
	}

	if err := filepath.Walk(dir, walcFunc); err != nil {
		return "", err
	}

	return ovfPath, nil
}

func calculateHash(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

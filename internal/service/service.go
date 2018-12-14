package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	vmware_types "github.com/vmware/govmomi/vim25/types"

	"github.com/vterdunov/janna-api/internal/config"
	"github.com/vterdunov/janna-api/internal/health"
	"github.com/vterdunov/janna-api/internal/types"
	"github.com/vterdunov/janna-api/internal/version"
)

// Service is the interface that represents methods of the business logic
type Service interface {
	GetConfig() *config.Config

	// Info returns some info about the Service
	Info() (string, string)

	// Healthz is a liveness probe
	Healthz() bool

	// Readyz is a readyness probe
	Readyz() bool

	// VMList returns list of VMs
	VMList(context.Context, *types.VMListParams) (map[string]string, error)

	// VMInfo provide summary information about VM
	VMInfo(context.Context, *types.VMInfoParams) (*VMSummary, error)

	// VMDelete destroys a Virtual Machine
	VMDelete(context.Context, *types.VMDeleteParams) error

	// VMFind find VM by path and return its UUID
	VMFind(context.Context, *types.VMFindParams) (*VMUuid, error)

	// VMDeploy create VM from OVA file
	VMDeploy(context.Context, *types.VMDeployParams) (string, error)

	// VMSnapshotsList returns VM snapshots list
	VMSnapshotsList(context.Context, *types.VMSnapshotsListParams) ([]Snapshot, error)

	// VMSnapshotCreate creates a VM snapshot
	VMSnapshotCreate(context.Context, *types.SnapshotCreateParams) (int32, error)

	// VMRestoreFromSnapshot creates a VM snapshot
	VMRestoreFromSnapshot(context.Context, *types.VMRestoreFromSnapshotParams) error

	// VMSnapshotDelete deletes snapshot
	VMSnapshotDelete(context.Context, *types.VMSnapshotDeleteParams) error

	VMPower(context.Context, *types.VMPowerParams) error

	VMRolesList(context.Context, *types.VMRolesListParams) ([]Role, error)

	VMAddRole(context.Context, *types.VMAddRoleParams) error

	VMScreenshot(context.Context, *types.VMScreenshotParams) ([]byte, error)

	RoleList(context.Context) ([]Role, error)

	// TasksList(context.Context) (*status.Tasks, error)

	TaskInfo(context.Context, string) (map[string]interface{}, error)

	// Reads Open API spec file
	OpenAPI(context.Context) ([]byte, error)
}

// service implements our Service
type service struct {
	logger   log.Logger
	cfg      *config.Config
	Client   *vim25.Client
	statuses Statuser
}

// New creates a new instance of the Service with wrapped middlewares
func New(
	logger log.Logger,
	cfg *config.Config,
	client *vim25.Client,
	duration metrics.Histogram,
	statuses Statuser,
) Service {
	// Build the layers of the service "onion" from the inside out.
	svc := NewSimpleService(logger, cfg, client, statuses)
	svc = NewLoggingService(log.With(logger, "component", "core"))(svc)
	svc = NewInstrumentingService(duration)(svc)

	return svc
}

// NewSimpleService creates a new instance of the Service with minimal preconfigured options
func NewSimpleService(
	logger log.Logger,
	cfg *config.Config,
	client *vim25.Client,
	statuses Statuser,
) Service {
	return &service{
		logger:   logger,
		cfg:      cfg,
		Client:   client,
		statuses: statuses,
	}
}

func (s *service) GetConfig() *config.Config {
	return s.cfg
}

func (s *service) Info() (string, string) {
	return version.GetBuildInfo()
}

func (s *service) Healthz() bool {
	return health.Healthz()
}

func (s *service) Readyz() bool {
	return health.Readyz()
}

func (s *service) VMList(ctx context.Context, params *types.VMListParams) (map[string]string, error) {
	root, err := chooseRoot(ctx, s.Client, params)
	if err != nil {
		return nil, err
	}

	m := view.NewManager(s.Client)
	v, err := m.CreateContainerView(ctx, root, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all machines
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)
	for i := range vms {
		vm := &vms[i]
		res[vm.Summary.Config.Uuid] = vm.Summary.Config.Name
	}

	return res, nil
}

func (s *service) VMInfo(ctx context.Context, params *types.VMInfoParams) (*VMSummary, error) {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	refs := make([]vmware_types.ManagedObjectReference, 0)
	refs = append(refs, vm.Reference())

	// Retrieve all properties
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var mVM mo.VirtualMachine
	var props []string

	pc := property.DefaultCollector(s.Client)

	if err := pc.Retrieve(ctx, refs, props, &mVM); err != nil {
		return nil, err
	}

	sum := VMSummary{
		Name: mVM.Summary.Config.Name,
	}

	return &sum, nil
}

func (s *service) VMDelete(ctx context.Context, params *types.VMDeleteParams) error {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	state, psErr := vm.PowerState(ctx)
	if psErr != nil {
		return errors.Wrap(psErr, "could not get Virtual Machine power state")
	}

	if state != vmware_types.VirtualMachinePowerStatePoweredOff {
		task, pOffErr := vm.PowerOff(ctx)
		if pOffErr != nil {
			return errors.Wrap(pOffErr, "could not power off Virtual Machine before destroying")
		}

		if err = task.Wait(ctx); err != nil {
			return errors.Wrap(err, "could not power off Virtual Machine before destroying")
		}
	}

	destroyTask, err := vm.Destroy(ctx)
	if err != nil {
		return err
	}

	return destroyTask.Wait(ctx)
}

func (s *service) VMFind(ctx context.Context, params *types.VMFindParams) (*VMUuid, error) {
	oVM, err := findByPath(ctx, s.Client, params.Datacenter, params.Path)
	if err != nil {
		return nil, err
	}

	refs := []vmware_types.ManagedObjectReference{oVM.Reference()}

	var vm mo.VirtualMachine

	pc := property.DefaultCollector(s.Client)

	if err := pc.Retrieve(ctx, refs, []string{"summary"}, &vm); err != nil {
		return nil, err
	}

	res := VMUuid{
		UUID: vm.Summary.Config.Uuid,
		Name: vm.Summary.Config.Name,
	}

	return &res, nil
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
		return "", fmt.Errorf("Virtual Machine '%s' already exist", params.Name) //nolint: golint
	}

	// do not import HTTP package. need interface or method to get request ID, or pass logger with reauest_id field
	// reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	reqID := ctx.Value("test-test-test-test")
	l := log.With(s.logger, "request_id", reqID)
	l = log.With(l, "vm", params.Name)

	taskCtx, cancel := context.WithTimeout(context.Background(), s.cfg.TaskTTL)

	t := s.statuses.NewTask()
	t.Str("stage", "start")

	// Start deploy in background
	go func() {
		defer cancel()
		d, err := newDeployment(taskCtx, s.Client, params, l, s.cfg)
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

func (s *service) VMSnapshotsList(ctx context.Context, params *types.VMSnapshotsListParams) ([]Snapshot, error) {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	return vmSnapshots(ctx, vm)
}

func (s *service) VMSnapshotCreate(ctx context.Context, params *types.SnapshotCreateParams) (int32, error) {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return 0, err
	}

	beforeSnapshots, err := vmSnapshots(ctx, vm)
	if err != nil {
		return 0, err
	}

	task, err := vm.CreateSnapshot(ctx, params.Name, params.Description, params.Memory, params.Quiesce)
	if err != nil {
		return 0, err
	}

	if errWait := task.Wait(ctx); errWait != nil {
		return 0, errWait
	}

	afterSnapshots, err := vmSnapshots(ctx, vm)
	if err != nil {
		return 0, err
	}

	afterIDs := make([]int32, 0, len(afterSnapshots))
	for _, i := range afterSnapshots {
		afterIDs = append(afterIDs, i.ID)
	}

	beforeIDs := make([]int32, 0, len(beforeSnapshots))
	for _, i := range beforeSnapshots {
		beforeIDs = append(beforeIDs, i.ID)
	}

	// at the same time somebody can create another snapshot. So, also, check snapshot names. I hope it enough.
	for _, i := range diff(afterIDs, beforeIDs) {
		for _, s := range afterSnapshots {
			if s.ID == i && s.Name == params.Name {
				return s.ID, nil
			}
		}
	}

	return 0, errors.New("could not get snapshot ID")
}

type snapshotReference struct {
	ref   vmware_types.ManagedObjectReference
	exist bool
}

func (s *snapshotReference) findByID(snapshots []vmware_types.VirtualMachineSnapshotTree, id int32) {
	for i := range snapshots {
		st := &snapshots[i]
		if id == st.Id {
			s.ref = st.Snapshot
			s.exist = true
		}
		s.findByID(st.ChildSnapshotList, id)
	}
}

func (s *snapshotReference) value() string {
	return s.ref.Value
}

func (s *service) VMRestoreFromSnapshot(ctx context.Context, params *types.VMRestoreFromSnapshotParams) error {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	var o mo.VirtualMachine

	err = vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return err
	}

	if o.Snapshot == nil || len(o.Snapshot.RootSnapshotList) == 0 {
		return errors.New("no snapshots for this VM")
	}

	sRef := &snapshotReference{}
	sRef.findByID(o.Snapshot.RootSnapshotList, params.SnapshotID)
	if !sRef.exist {
		return fmt.Errorf("cound not find snapshot with id %d", params.SnapshotID)
	}

	task, err := vm.RevertToSnapshot(ctx, sRef.value(), params.PowerOn)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

func (s *service) VMSnapshotDelete(ctx context.Context, params *types.VMSnapshotDeleteParams) error {
	// snapshot lookup by name, where name can be:
	// 1) snapshot ManagedObjectReference.Value (unique)
	// 2) snapshot name (may not be unique)
	// 3) snapshot tree path (may not be unique)
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	var o mo.VirtualMachine

	err = vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return err
	}

	if o.Snapshot == nil || len(o.Snapshot.RootSnapshotList) == 0 {
		return errors.New("no snapshots for this VM")
	}

	sRef := &snapshotReference{}
	sRef.findByID(o.Snapshot.RootSnapshotList, params.SnapshotID)
	if !sRef.exist {
		return fmt.Errorf("cound not find snapshot with id %d", params.SnapshotID)
	}

	task, err := vm.RemoveSnapshot(ctx, sRef.value(), false, nil)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

// Power changes VM power state
func (s *service) VMPower(ctx context.Context, params *types.VMPowerParams) error {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	switch params.State {
	case "on":
		err = powerOn(ctx, vm)
	case "off":
		err = powerOff(ctx, vm)
	case "suspend":
		err = suspend(ctx, vm)
	case "reboot":
		err = reboot(ctx, vm)
	case "reset":
		err = reset(ctx, vm)
	}

	return err
}

func (s *service) VMRolesList(ctx context.Context, params *types.VMRolesListParams) ([]Role, error) {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	am := object.NewAuthorizationManager(s.Client)

	perms, err := am.RetrieveEntityPermissions(ctx, vm.Reference(), true)
	if err != nil {
		return nil, err
	}

	for _, p := range perms {
		_ = p
		// fmt.Println(p.Principal)
	}

	roles, err := am.RoleList(ctx)
	if err != nil {
		return nil, err
	}

	rr := []Role{}
	for _, role := range roles {
		desc := role.Info.GetDescription()
		r := Role{
			Name: role.Name,
			ID:   role.RoleId,
		}

		r.Description.Label = desc.Label
		r.Description.Summary = desc.Summary
		rr = append(rr, r)
	}

	// TODO: Implement get role name from IDs
	return rr, nil
}

func (s *service) VMAddRole(ctx context.Context, params *types.VMAddRoleParams) error {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	p := vmware_types.Permission{
		Principal: params.Principal,
		RoleId:    params.RoleID,
	}
	pp := []vmware_types.Permission{}
	pp = append(pp, p)

	am := object.NewAuthorizationManager(s.Client)
	if err := am.SetEntityPermissions(ctx, vm.Reference(), pp); err != nil {
		return err
	}

	return nil
}

func (s *service) RoleList(ctx context.Context) ([]Role, error) {
	am := object.NewAuthorizationManager(s.Client)
	roles, err := am.RoleList(ctx)
	if err != nil {
		return nil, err
	}

	rr := []Role{}
	for _, role := range roles {
		desc := role.Info.GetDescription()
		r := Role{
			Name: role.Name,
			ID:   role.RoleId,
		}

		r.Description.Label = desc.Label
		r.Description.Summary = desc.Summary
		rr = append(rr, r)
	}

	return rr, err
}

func (s *service) VMScreenshot(ctx context.Context, params *types.VMScreenshotParams) ([]byte, error) {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	state, psErr := vm.PowerState(ctx)
	if psErr != nil {
		return nil, errors.Wrap(psErr, "could not get Virtual Machine power state")
	}

	if state != vmware_types.VirtualMachinePowerStatePoweredOn {
		return nil, fmt.Errorf("vm is not powered on (%s)", state)
	}

	u := s.Client.URL()
	u.Path = "/screen"
	query := url.Values{"id": []string{vm.Reference().Value}}
	u.RawQuery = query.Encode()

	param := soap.DefaultDownload

	rc, _, derr := s.Client.Download(ctx, u, &param)
	if derr != nil {
		return nil, derr
	}
	defer rc.Close()

	screenshot, rErr := ioutil.ReadAll(rc)
	if rErr != nil {
		return nil, rErr
	}

	return screenshot, nil
}

func (s *service) TaskInfo(ctx context.Context, taskID string) (map[string]interface{}, error) {
	t := s.statuses.FindByID(taskID)
	if t != nil {
		return t.Get(), nil
	}
	return nil, errors.New("task not found")
}

func (s *service) OpenAPI(_ context.Context) ([]byte, error) {
	spec, err := ioutil.ReadFile("./api/openapi.json")
	if err != nil {
		return nil, err
	}
	return spec, err
}

// findByUUID find and returns VM by its UUID
func findByUUID(ctx context.Context, client *vim25.Client, DCName, uuid string) (*object.VirtualMachine, error) {
	f := find.NewFinder(client, true)

	dc, err := f.DatacenterOrDefault(ctx, DCName)
	if err != nil {
		return nil, err
	}

	f.SetDatacenter(dc)

	si := object.NewSearchIndex(client)

	ref, err := si.FindByUuid(ctx, dc, uuid, true, nil)
	if err != nil {
		return nil, err
	}

	vm, ok := ref.(*object.VirtualMachine)
	if !ok {
		return nil, errors.New("could not find Virtual Machine by UUID. Could not assert reference to Virtual Machine")
	}

	return vm, nil
}

func chooseRoot(ctx context.Context, c *vim25.Client, params *types.VMListParams) (vmware_types.ManagedObjectReference, error) {
	var ref vmware_types.ManagedObjectReference
	f := find.NewFinder(c, true)
	dc, err := f.DatacenterOrDefault(ctx, params.Datacenter)
	if err != nil {
		return ref, err
	}

	if params.Folder != "" {
		f.SetDatacenter(dc)
		rp, err := f.FolderOrDefault(ctx, params.Folder)
		if err != nil {
			return ref, err
		}
		return rp.Reference(), nil
	}

	if params.ResourcePool != "" {
		f.SetDatacenter(dc)
		rp, err := f.ResourcePoolOrDefault(ctx, params.ResourcePool)
		if err != nil {
			return ref, err
		}
		return rp.Reference(), nil
	}
	return dc.Reference(), nil
}

// findByPath find and returns VM by Inventory Path
func findByPath(ctx context.Context, client *vim25.Client, DCname, path string) (*object.VirtualMachine, error) {
	f := find.NewFinder(client, true)

	dc, err := f.DatacenterOrDefault(ctx, DCname)
	if err != nil {
		return nil, err
	}

	f.SetDatacenter(dc)

	return f.VirtualMachine(ctx, path)
}

func vmSnapshots(ctx context.Context, vm *object.VirtualMachine) ([]Snapshot, error) {
	var o mo.VirtualMachine

	err := vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return nil, err
	}

	st := make([]Snapshot, 0)
	if o.Snapshot == nil {
		return st, nil
	}

	ch := make(chan Snapshot, 1000)
	walk(o.Snapshot.RootSnapshotList, ch)

	close(ch)
	for v := range ch {
		st = append(st, v)
	}

	return st, nil
}

func walk(st []vmware_types.VirtualMachineSnapshotTree, ch chan Snapshot) {
	for i := range st {
		s := &st[i]
		t := Snapshot{
			Name:        s.Name,
			ID:          s.Id,
			Description: s.Description,
			CreatedAt:   s.CreateTime,
		}

		ch <- t
		walk(s.ChildSnapshotList, ch)
	}
}

func diff(slice1 []int32, slice2 []int32) []int32 {
	var diff []int32

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

// Power

const (
	off       = vmware_types.VirtualMachinePowerStatePoweredOff
	on        = vmware_types.VirtualMachinePowerStatePoweredOn
	suspended = vmware_types.VirtualMachinePowerStateSuspended
)

// powerOn power on Virtual Machine
func powerOn(ctx context.Context, vm *object.VirtualMachine) error {
	state, err := getVMPowerState(ctx, vm)
	if err != nil {
		return err
	}

	switch state {
	case on:
		return nil

	case off, suspended:
		task, err := vm.PowerOn(ctx)
		if err != nil {
			return err
		}

		return task.Wait(ctx)

	default:
		// actually there are no more states. safe to the future.
		return fmt.Errorf("could not power on Virtual Machine from %s state", state)
	}
}

// powerOff power off Virtual Machine
func powerOff(ctx context.Context, vm *object.VirtualMachine) error {
	state, err := getVMPowerState(ctx, vm)
	if err != nil {
		return err
	}

	switch state {
	case off:
		return nil

	case on:
		err := vm.ShutdownGuest(ctx)
		if err != nil && isToolsUnavailable(err) {
			task, powerErr := vm.PowerOff(ctx)
			if powerErr != nil {
				return powerErr
			}

			return task.Wait(ctx)
		}
		return err

	case suspended:
		task, err := vm.PowerOff(ctx)
		if err != nil {
			return err
		}

		return task.Wait(ctx)

	default:
		return fmt.Errorf("could not power off Virtual Machine from %s state", state)
	}
}

// reboot Virtual Machine. It tries to use VMWareTools to call guest agent to reboot the VM.
// And as the last way, the method tries to reset VM.
func reboot(ctx context.Context, vm *object.VirtualMachine) error {
	state, err := getVMPowerState(ctx, vm)
	if err != nil {
		return err
	}

	if state != on {
		return fmt.Errorf("could not reboot Virtual Machine from %s state", state)
	}

	err = vm.RebootGuest(ctx)
	if err != nil && isToolsUnavailable(err) {
		task, resetErr := vm.Reset(ctx)
		if resetErr != nil {
			return resetErr
		}

		return task.Wait(ctx)
	}

	return err
}

// reset Virtual Machine
func reset(ctx context.Context, vm *object.VirtualMachine) error {
	state, err := getVMPowerState(ctx, vm)
	if err != nil {
		return err
	}

	if state != on {
		return fmt.Errorf("could not reset Virtual Machine from %s state", state)
	}

	task, err := vm.Reset(ctx)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

// suspend Virtual Machine
func suspend(ctx context.Context, vm *object.VirtualMachine) error {
	state, err := getVMPowerState(ctx, vm)
	if err != nil {
		return err
	}

	switch state {
	case suspended:
		return nil

	case on:
		task, err := vm.Suspend(ctx)
		if err != nil {
			return err
		}

		return task.Wait(ctx)

	default:
		return fmt.Errorf("could not suspend Virtual Machine from %s state", state)
	}
}

func getVMPowerState(ctx context.Context, vm *object.VirtualMachine) (vmware_types.VirtualMachinePowerState, error) {
	state, err := vm.PowerState(ctx)
	if err != nil {
		return "", err
	}

	return state, err
}

func isToolsUnavailable(err error) bool {
	if soap.IsSoapFault(err) {
		soapFault := soap.ToSoapFault(err)
		if _, ok := soapFault.VimFault().(vmware_types.ToolsUnavailable); ok {
			return ok
		}
	}

	return false
}

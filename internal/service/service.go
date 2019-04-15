// service provides use-cases for the Service
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
	"github.com/vterdunov/janna-api/internal/domain"
	"github.com/vterdunov/janna-api/internal/health"
	"github.com/vterdunov/janna-api/internal/types"
	"github.com/vterdunov/janna-api/internal/version"
)

type contextKey int

const ContextKeyRequestXRequestID contextKey = iota

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
	VMList(context.Context, *types.VMListParams) ([]domain.VMUuid, error)

	// VMInfo provide summary information about VM
	VMInfo(context.Context, *types.VMInfoParams) (*domain.VMSummary, error)

	// VMDelete destroys a Virtual Machine
	VMDelete(context.Context, *types.VMDeleteParams) error

	// VMFind find VM by path and return its UUID
	VMFind(context.Context, *types.VMFindParams) (*domain.VMUuid, error)

	// VMDeploy create VM from OVA file
	VMDeploy(context.Context, *types.VMDeployParams) (string, error)

	// VMSnapshotsList returns VM snapshots list
	VMSnapshotsList(context.Context, *types.VMSnapshotsListParams) ([]domain.Snapshot, error)

	// VMSnapshotCreate creates a VM snapshot
	VMSnapshotCreate(context.Context, *types.SnapshotCreateParams) (int32, error)

	// VMRestoreFromSnapshot creates a VM snapshot
	VMRestoreFromSnapshot(context.Context, *types.VMRestoreFromSnapshotParams) error

	// VMSnapshotDelete deletes snapshot
	VMSnapshotDelete(context.Context, *types.VMSnapshotDeleteParams) error

	VMPower(context.Context, *types.VMPowerParams) error

	VMRolesList(context.Context, *types.VMRolesListParams) ([]domain.Role, error)

	VMAddRole(context.Context, *types.VMAddRoleParams) error

	VMScreenshot(context.Context, *types.VMScreenshotParams) ([]byte, error)

	RoleList(context.Context) ([]domain.Role, error)

	// TasksList(context.Context) (*status.Tasks, error)

	TaskInfo(context.Context, string) (map[string]interface{}, error)

	// Reads Open API spec file
	OpenAPI(context.Context) ([]byte, error)

	// VMRename renames Virtual Machine
	VMRename(context.Context, *types.VMRenameParams) error
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

func (s *service) VMList(ctx context.Context, params *types.VMListParams) ([]domain.VMUuid, error) {
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

	resVMs := []domain.VMUuid{}
	for i := range vms {
		vm := &vms[i]
		vmUUID := domain.VMUuid{
			Name: vm.Summary.Config.Name,
			UUID: vm.Summary.Config.Uuid,
		}
		resVMs = append(resVMs, vmUUID)
	}

	return resVMs, nil
}

func (s *service) VMInfo(ctx context.Context, params *types.VMInfoParams) (*domain.VMSummary, error) {
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

	gi := domain.VMGuestInfo{
		GuestID:            mVM.Summary.Guest.GuestId,
		GuestFullName:      mVM.Summary.Guest.GuestFullName,
		ToolsRunningStatus: mVM.Summary.Guest.ToolsRunningStatus,
		HostName:           mVM.Summary.Guest.HostName,
		IPAddress:          mVM.Summary.Guest.IpAddress,
	}

	sum := domain.VMSummary{
		Name:             mVM.Summary.Config.Name,
		UUID:             mVM.Summary.Config.Uuid,
		Template:         mVM.Summary.Config.Template,
		GuestID:          mVM.Summary.Config.GuestId,
		Annotation:       mVM.Summary.Config.Annotation,
		PowerState:       string(mVM.Runtime.PowerState),
		NumCPU:           mVM.Summary.Config.NumCpu,
		NumEthernetCards: mVM.Summary.Config.NumEthernetCards,
		NumVirtualDisks:  mVM.Summary.Config.NumVirtualDisks,
		VMGuestInfo:      gi,
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

func (s *service) VMFind(ctx context.Context, params *types.VMFindParams) (*domain.VMUuid, error) {
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

	res := domain.VMUuid{
		UUID: vm.Summary.Config.Uuid,
		Name: vm.Summary.Config.Name,
	}

	return &res, nil
}

func (s *service) VMRolesList(ctx context.Context, params *types.VMRolesListParams) ([]domain.Role, error) {
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
	}

	roles, err := am.RoleList(ctx)
	if err != nil {
		return nil, err
	}

	rr := []domain.Role{}
	for _, role := range roles {
		desc := role.Info.GetDescription()
		r := domain.Role{
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

func (s *service) RoleList(ctx context.Context) ([]domain.Role, error) {
	am := object.NewAuthorizationManager(s.Client)
	roles, err := am.RoleList(ctx)
	if err != nil {
		return nil, err
	}

	rr := []domain.Role{}
	for _, role := range roles {
		desc := role.Info.GetDescription()
		r := domain.Role{
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

func (s *service) VMRename(ctx context.Context, params *types.VMRenameParams) error {
	f := NewFinder(params.Datacenter, s.Client)

	vm, err := f.FindVMByUUID(params.UUID)
	if err != nil {
		return err
	}

	return vm.Rename(ctx, params.Name)
}

func newWithObjectVM(vmwareVM *object.VirtualMachine) *domain.VirtualMachine {
	return &domain.VirtualMachine{
		VMWareVM: vmwareVM,
	}
}

type Finder struct {
	dc     string
	client *vim25.Client
}

func NewFinder(dc string, c *vim25.Client) Finder {
	return Finder{
		dc:     dc,
		client: c,
	}
}

func (f *Finder) FindVMByUUID(uuid string) (*domain.VirtualMachine, error) {
	vmwareFinder := find.NewFinder(f.client, true)

	ctx := context.TODO()
	dc, err := vmwareFinder.DatacenterOrDefault(ctx, f.dc)
	if err != nil {
		return nil, err
	}

	vmwareFinder.SetDatacenter(dc)

	si := object.NewSearchIndex(f.client)

	ref, err := si.FindByUuid(ctx, dc, uuid, true, nil)
	if err != nil {
		return nil, err
	}

	vm, ok := ref.(*object.VirtualMachine)
	if !ok {
		return nil, errors.New("could not find Virtual Machine by UUID. Could not assert reference to Virtual Machine")
	}

	return newWithObjectVM(vm), nil
}

// findByUUID find and returns VM by its UUID
func findByUUID(ctx context.Context, client *vim25.Client, dcName, uuid string) (*object.VirtualMachine, error) {
	f := find.NewFinder(client, true)

	dc, err := f.DatacenterOrDefault(ctx, dcName)
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
func findByPath(ctx context.Context, client *vim25.Client, dcName, path string) (*object.VirtualMachine, error) {
	f := find.NewFinder(client, true)

	dc, err := f.DatacenterOrDefault(ctx, dcName)
	if err != nil {
		return nil, err
	}

	f.SetDatacenter(dc)

	return f.VirtualMachine(ctx, path)
}

func vmSnapshots(ctx context.Context, vm *object.VirtualMachine) ([]domain.Snapshot, error) {
	var o mo.VirtualMachine

	err := vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return nil, err
	}

	st := make([]domain.Snapshot, 0)
	if o.Snapshot == nil {
		return st, nil
	}

	ch := make(chan domain.Snapshot, 1000)
	walk(o.Snapshot.RootSnapshotList, ch)

	close(ch)
	for v := range ch {
		st = append(st, v)
	}

	return st, nil
}

func walk(st []vmware_types.VirtualMachineSnapshotTree, ch chan domain.Snapshot) {
	for i := range st {
		s := &st[i]
		t := domain.Snapshot{
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

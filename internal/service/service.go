package service

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"

	"github.com/vterdunov/janna-api/internal/config"
	"github.com/vterdunov/janna-api/internal/health"
	"github.com/vterdunov/janna-api/internal/providers/vmware/permissions"
	"github.com/vterdunov/janna-api/internal/providers/vmware/vm"
	"github.com/vterdunov/janna-api/internal/types"
	"github.com/vterdunov/janna-api/internal/version"
	"github.com/vterdunov/janna-api/pkg/uuid"
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
	VMInfo(context.Context, *types.VMInfoParams) (*types.VMSummary, error)

	// VMDelete destroys a Virtual Machine
	VMDelete(context.Context, *types.VMDeleteParams) error

	// VMFind find VM by path and return its UUID
	VMFind(context.Context, *types.VMFindParams) (*types.VMFound, error)

	// VMDeploy create VM from OVA file
	VMDeploy(context.Context, *types.VMDeployParams) (string, error)

	// VMSnapshotsList returns VM snapshots list
	VMSnapshotsList(context.Context, *types.VMSnapshotsListParams) ([]types.Snapshot, error)

	// VMSnapshotCreate creates a VM snapshot
	VMSnapshotCreate(context.Context, *types.SnapshotCreateParams) (int32, error)

	// VMRestoreFromSnapshot creates a VM snapshot
	VMRestoreFromSnapshot(context.Context, *types.VMRestoreFromSnapshotParams) error

	// VMSnapshotDelete deletes snapshot
	VMSnapshotDelete(context.Context, *types.VMSnapshotDeleteParams) error

	VMRolesList(context.Context, *types.VMRolesListParams) ([]types.Role, error)

	VMAddRole(context.Context, *types.VMAddRoleParams) error

	RoleList(context.Context) ([]types.Role, error)

	// TasksList(context.Context) (*status.Tasks, error)

	TaskInfo(context.Context, string) (*Task, error)

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
	svc := newSimpleService(logger, cfg, client, statuses)
	svc = NewLoggingService(log.With(logger, "component", "core"))(svc)
	svc = NewInstrumentingService(duration)(svc)

	return svc
}

// newSimpleService creates a new instance of the Service with minimal preconfigured options
func newSimpleService(
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
	return vm.List(ctx, s.Client, params)
}

func (s *service) VMInfo(ctx context.Context, params *types.VMInfoParams) (*types.VMSummary, error) {
	return vm.Info(ctx, s.Client, params)
}

func (s *service) VMDelete(ctx context.Context, params *types.VMDeleteParams) error {
	return vm.Delete(ctx, s.Client, params)
}

func (s *service) VMFind(ctx context.Context, params *types.VMFindParams) (*types.VMFound, error) {
	return vm.Find(ctx, s.Client, params)
}

func (s *service) VMDeploy(ctx context.Context, params *types.VMDeployParams) (string, error) {
	// TODO: validate incoming params according business rules (https://github.com/asaskevich/govalidator)
	// use Endpoint middleware

	// predeploy checks
	exist, err := vm.IsVMExist(ctx, s.Client, params)
	if err != nil {
		return "", err
	}

	if exist {
		return "", fmt.Errorf("Virtual Machine '%s' already exist", params.Name) // nolint: golint
	}

	taskID := uuid.NewUUID()
	s.statuses.Add(taskID, "Start deploy")

	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	l := log.With(s.logger, "request_id", reqID)
	l = log.With(l, "vm", params.Name)

	taskCtx, cancel := context.WithTimeout(context.Background(), s.cfg.TaskTTL)

	// Start deploy in background
	go func() {
		defer cancel()
		d, err := vm.NewDeployment(taskCtx, s.Client, params, l, s.cfg)
		if err != nil {
			err = errors.Wrap(err, "Could not create deployment object")
			l.Log("err", err)
			s.statuses.Add(taskID, err.Error())
			cancel()
			return
		}

		s.statuses.Add(taskID, "Importing OVA")
		moref, err := d.Import(taskCtx, params.OVAURL, params.Annotation)
		if err != nil {
			err = errors.Wrap(err, "Could not import OVA/OVF")
			l.Log("err", err)
			s.statuses.Add(taskID, err.Error())
			cancel()
			return
		}

		s.statuses.Add(taskID, "Creating Virtual Machine")
		vmx := object.NewVirtualMachine(s.Client, *moref)

		l.Log("msg", "Powering on...")
		s.statuses.Add(taskID, "Powering on")
		if err = vm.PowerON(taskCtx, vmx); err != nil {
			err = errors.Wrap(err, "Could not Virtual Machine power on")
			l.Log("err", err)
			s.statuses.Add(taskID, err.Error())
			cancel()
			return
		}

		s.statuses.Add(taskID, "Waiting for IP")
		ip, err := vm.WaitForIP(taskCtx, vmx)
		if err != nil {
			err = errors.Wrap(err, "error getting IP address")
			l.Log("err", err)
			s.statuses.Add(taskID, err.Error())
			cancel()
			return
		}

		l.Log("msg", "Successful deploy", "ip", ip)
		s.statuses.Add(taskID, fmt.Sprintf("Done, IP: %s", ip))
	}()

	return taskID, nil
}

func (s *service) VMSnapshotsList(ctx context.Context, params *types.VMSnapshotsListParams) ([]types.Snapshot, error) {
	st, err := vm.SnapshotsList(ctx, s.Client, params)
	if err != nil {
		return nil, err
	}

	return st, nil
}

func (s *service) VMSnapshotCreate(ctx context.Context, params *types.SnapshotCreateParams) (int32, error) {
	return vm.SnapshotCreate(ctx, s.Client, params)
}

func (s *service) VMRestoreFromSnapshot(ctx context.Context, params *types.VMRestoreFromSnapshotParams) error {
	return vm.RestoreFromSnapshot(ctx, s.Client, params)
}

func (s *service) VMSnapshotDelete(ctx context.Context, params *types.VMSnapshotDeleteParams) error {
	return vm.DeleteSnapshot(ctx, s.Client, params)
}

func (s *service) VMRolesList(ctx context.Context, params *types.VMRolesListParams) ([]types.Role, error) {
	return vm.RolesList(ctx, s.Client, params)
}

func (s *service) VMAddRole(ctx context.Context, params *types.VMAddRoleParams) error {
	return vm.AddRole(ctx, s.Client, params)
}

func (s *service) RoleList(ctx context.Context) ([]types.Role, error) {
	return permissions.RoleList(ctx, s.Client)
}

func (s *service) TaskInfo(ctx context.Context, taskID string) (*Task, error) {
	t := s.statuses.Get(taskID)
	if t != nil {
		return t, nil
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

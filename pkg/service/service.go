package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/vterdunov/janna-api/pkg/status"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/vmware/govmomi/vim25"

	"github.com/vterdunov/janna-api/pkg/config"
	"github.com/vterdunov/janna-api/pkg/health"
	"github.com/vterdunov/janna-api/pkg/providers/vmware/permissions"
	"github.com/vterdunov/janna-api/pkg/providers/vmware/vm"
	"github.com/vterdunov/janna-api/pkg/types"
	"github.com/vterdunov/janna-api/pkg/uuid"
	"github.com/vterdunov/janna-api/pkg/version"
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

	// VMFind find VM by path and return its UUID
	VMFind(context.Context, *types.VMFindParams) (*types.VMFound, error)

	// VMDeploy create VM from OVA file
	VMDeploy(context.Context, *types.VMDeployParams) (int, error)

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

	TaskInfo(context.Context, string) (*status.Task, error)
}

// service implements our Service
type service struct {
	logger   log.Logger
	cfg      *config.Config
	Client   *vim25.Client
	statuses *status.Tasks
}

// New creates a new instance of the Service with wrapped middlewares
func New(logger log.Logger, cfg *config.Config, client *vim25.Client, duration metrics.Histogram) Service {
	svc := NewSimpleService(logger, cfg, client)
	svc = NewLoggingService(log.With(logger, "component", "core"))(svc)
	svc = NewInstrumentingService(duration)(svc)

	return svc
}

// NewSimpleService creates a new instance of the Service with minimal preconfigured options
func NewSimpleService(logger log.Logger, cfg *config.Config, client *vim25.Client) Service {
	statuses := status.New()
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

func (s *service) VMFind(ctx context.Context, params *types.VMFindParams) (*types.VMFound, error) {
	return vm.Find(ctx, s.Client, params)
}

func (s *service) VMDeploy(ctx context.Context, params *types.VMDeployParams) (int, error) {
	// TODO: validate incoming params according business rules (https://github.com/asaskevich/govalidator)
	taskID := uuid.NewUUID()
	s.statuses.Add(taskID, "Start deploy")
	status := s.statuses.Get(taskID)
	if status != nil {
		// fmt.Println(status.Status)
		fmt.Println(taskID)
	}

	return vm.Deploy(ctx, s.Client, params, s.logger, s.cfg)
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

func (s *service) TaskInfo(ctx context.Context, taskID string) (*status.Task, error) {
	t := s.statuses.Get(taskID)
	if t != nil {
		return t, nil
	}
	return nil, errors.New("Not implemented")
}

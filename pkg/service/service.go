package service

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/vmware/govmomi/vim25"

	"github.com/vterdunov/janna-api/pkg/config"
	"github.com/vterdunov/janna-api/pkg/health"
	"github.com/vterdunov/janna-api/pkg/providers/vmware/vm"
	"github.com/vterdunov/janna-api/pkg/types"
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
	VMList(context.Context, string) ([]string, error)

	// VMInfo provide summary information about VM
	VMInfo(context.Context, *types.VMInfoParams) (*types.VMSummary, error)

	// VMDeploy create VM from OVA file
	VMDeploy(context.Context, *types.VMDeployParams) (int, error)

	// VMSnapshotsList returns VM snapshots list
	VMSnapshotsList(context.Context, *types.VMSnapshotsListParams) ([]types.Snapshot, error)

	// VMSnapshotCreate creates a VM snapshot
	VMSnapshotCreate(context.Context, *types.SnapshotCreateParams) error

	// VMRestoreFromSnapshot creates a VM snapshot
	VMRestoreFromSnapshot(context.Context, *types.VMRestoreFromSnapshotParams) error
}

// service implements our Service
type service struct {
	logger log.Logger
	cfg    *config.Config
	Client *vim25.Client
}

// New creates a new instance of the Service with some preconfigured options
func New(logger log.Logger, cfg *config.Config, client *vim25.Client) Service {
	return service{
		logger: log.With(logger, "component", "core"),
		cfg:    cfg,
		Client: client,
	}
}

func (s service) GetConfig() *config.Config {
	return s.cfg
}

func (s service) Info() (string, string) {
	return version.GetBuildInfo()
}

func (s service) Healthz() bool {
	return health.Healthz()
}

func (s service) Readyz() bool {
	return health.Readyz()
}

func (s service) VMList(ctx context.Context, folder string) ([]string, error) {
	// TODO: Implement business logic
	var vms []string
	return vms, nil
}

func (s service) VMInfo(ctx context.Context, params *types.VMInfoParams) (*types.VMSummary, error) {
	return vm.Info(ctx, s.Client, params)
}

func (s service) VMDeploy(ctx context.Context, params *types.VMDeployParams) (int, error) {
	// TODO: validate incoming params according business rules (https://github.com/asaskevich/govalidator)

	return vm.Deploy(ctx, s.Client, params, s.logger, s.cfg)
}

func (s service) VMSnapshotsList(ctx context.Context, params *types.VMSnapshotsListParams) ([]types.Snapshot, error) {
	st, err := vm.SnapshotsList(ctx, s.Client, params)
	if err != nil {
		return nil, err
	}

	return st, nil
}

func (s service) VMSnapshotCreate(ctx context.Context, params *types.SnapshotCreateParams) error {
	return vm.SnapshotCreate(ctx, s.Client, params)
}

func (s service) VMRestoreFromSnapshot(ctx context.Context, params *types.VMRestoreFromSnapshotParams) error {
	return vm.RestoreFromSnapshot(ctx, s.Client, params)
}

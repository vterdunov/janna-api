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
	// Info returns some info about the Service
	Info() (string, string)

	// Healthz is a liveness probe
	Healthz() bool

	// Readyz is a readyness probe
	Readyz() bool

	// VMList returns list of VMs
	VMList(context.Context, string) ([]string, error)

	// VMInfo provide summary information about VM
	VMInfo(context.Context, string) (*types.VMSummary, error)

	// VMDeploy create VM from OVA file
	VMDeploy(context.Context, *types.VMDeployParams) (int, error)

	// VMSnapshotsList returns VM snapshots list
	VMSnapshotsList(context.Context, string) ([]string, error)
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

func (s service) VMInfo(ctx context.Context, name string) (*types.VMSummary, error) {
	return vm.Info(ctx, name, s.logger, s.cfg, s.Client)
}

func (s service) VMDeploy(ctx context.Context, deployParams *types.VMDeployParams) (int, error) {
	// TODO: validate incoming params according business rules (https://github.com/asaskevich/govalidator)

	return vm.Deploy(ctx, deployParams, s.logger, s.cfg, s.Client)
}

func (s service) VMSnapshotsList(ctx context.Context, vmName string) ([]string, error) {
	st, err := vm.VMSnapshotsList(ctx, s.Client, s.cfg, vmName)
	if err != nil {
		return nil, err
	}

	return st, nil
}

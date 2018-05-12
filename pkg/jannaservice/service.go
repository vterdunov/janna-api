package jannaservice

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

	// VMInfo provide summary information about VM
	VMInfo(context.Context, string) (*types.VMSummary, error)

	// VMDeploy create VM from OVA file
	VMDeploy(context.Context, *types.VMDeployParams) (int, error)
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

func (s service) VMInfo(ctx context.Context, name string) (*types.VMSummary, error) {
	return vm.Info(ctx, name, s.logger, s.cfg, s.Client)
}

func (s service) VMDeploy(ctx context.Context, deployParams *types.VMDeployParams) (int, error) {
	return vm.Deploy(ctx, deployParams, s.logger, s.cfg, s.Client)
}

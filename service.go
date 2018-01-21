package main

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/vterdunov/janna-api/health"

	"github.com/vterdunov/janna-api/providers/vmware/vm"
	"github.com/vterdunov/janna-api/types"
	"github.com/vterdunov/janna-api/version"
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
	VMInfo(context.Context, string) (types.VMSummary, error)
}

// service implements our Service
type service struct {
	logger log.Logger
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

func (s service) VMInfo(ctx context.Context, name string) (types.VMSummary, error) {
	return vm.Info(ctx, name)
}

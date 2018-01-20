package main

import (
	"context"

	"github.com/vterdunov/janna-api/providers/vmware/vm"
	"github.com/vterdunov/janna-api/types"
)

// Service is the interface that represents methods of the business logic
type Service interface {
	// VMInfo provide summary information about VM
	VMInfo(context.Context, string) (types.VMSummary, error)
}

// service implements our Service
type service struct{}

func (s service) VMInfo(ctx context.Context, name string) (types.VMSummary, error) {
	return vm.Info(ctx, name)
}

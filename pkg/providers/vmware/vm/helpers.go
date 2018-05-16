package vm

import (
	"context"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"

	"github.com/vterdunov/janna-api/pkg/config"
)

// FindByUUID find and returns VM by its UUID
func FindByUUID(ctx context.Context, client *vim25.Client, cfg *config.Config, uuid string) (*object.VirtualMachine, error) {
	f := find.NewFinder(client, true)

	dc, err := f.DatacenterOrDefault(ctx, cfg.VMWare.DC)
	if err != nil {
		return nil, err
	}

	f.SetDatacenter(dc)

	return f.VirtualMachine(ctx, uuid)
}

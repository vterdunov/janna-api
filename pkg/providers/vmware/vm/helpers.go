package vm

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
)

// FindByUUID find and returns VM by its UUID
func FindByUUID(ctx context.Context, client *vim25.Client, DCName, uuid string) (*object.VirtualMachine, error) {
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
		return nil, errors.New("Could not find Virtual Machine by UUID. Could not assert reference to Virtual Machine")
	}

	return vm, nil
}

// FindByPath find and returns VM by Inventory Path
func FindByPath(ctx context.Context, client *vim25.Client, DCname, path string) (*object.VirtualMachine, error) {
	f := find.NewFinder(client, true)

	dc, err := f.DatacenterOrDefault(ctx, DCname)
	if err != nil {
		return nil, err
	}

	f.SetDatacenter(dc)

	return f.VirtualMachine(ctx, path)
}

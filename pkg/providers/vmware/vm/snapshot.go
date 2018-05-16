package vm

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"

	"github.com/vterdunov/janna-api/pkg/config"
)

// VMSnapshotsList return a list of VM snapshots
func VMSnapshotsList(ctx context.Context, client *vim25.Client, cfg *config.Config, vmName string) ([]string, error) {
	var st []string

	vm, err := FindByUUID(ctx, client, cfg, vmName)
	if err != nil {
		return nil, err
	}

	var o mo.VirtualMachine

	err = vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return nil, err
	}

	if o.Snapshot == nil {
		return nil, err
	}
	for _, s := range o.Snapshot.RootSnapshotList {
		fmt.Println(s.Name)
		fmt.Println(s.Description)
		fmt.Println(s.Id)

		st = append(st, s.Name)
	}

	return st, nil
}

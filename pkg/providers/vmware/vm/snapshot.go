package vm

import (
	"context"

	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/pkg/types"
)

// SnapshotsList return a list of VM snapshots
func SnapshotsList(ctx context.Context, client *vim25.Client, params *jt.VMSnapshotsListParams) ([]jt.Snapshot, error) {
	st := make([]jt.Snapshot, 0)

	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
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

	ch := make(chan jt.Snapshot, 1000)
	walk(o.Snapshot.RootSnapshotList, ch)

	close(ch)
	for v := range ch {
		st = append(st, v)
	}

	return st, nil
}

// SnapshotCreate creates VM snapshot
func SnapshotCreate(ctx context.Context, client *vim25.Client, params *jt.SnapshotCreateParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	task, err := vm.CreateSnapshot(ctx, params.Name, params.Description, params.Memory, params.Quiesce)
	if err != nil {
		return err
	}

	if err := task.Wait(ctx); err != nil {
		return err
	}

	return nil
}

// RestoreFromSnapshot creates VM snapshot
func RestoreFromSnapshot(ctx context.Context, client *vim25.Client, params *jt.VMRestoreFromSnapshotParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	task, err := vm.RevertToSnapshot(ctx, params.Name, params.PowerOn)
	if err != nil {
		return err
	}

	if err := task.Wait(ctx); err != nil {
		return err
	}

	return nil
}

func walk(st []types.VirtualMachineSnapshotTree, ch chan jt.Snapshot) {
	for _, s := range st {
		t := jt.Snapshot{
			Name:        s.Name,
			ID:          s.Id,
			Description: s.Description,
			CreatedAt:   s.CreateTime,
		}

		ch <- t
		walk(s.ChildSnapshotList, ch)
	}
}

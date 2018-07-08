package vm

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/pkg/types"
)

type snapshotReference struct {
	ref   types.ManagedObjectReference
	exist bool
}

// SnapshotsList return a list of VM snapshots
func SnapshotsList(ctx context.Context, client *vim25.Client, params *jt.VMSnapshotsListParams) ([]jt.Snapshot, error) {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	return vmSnapshots(ctx, vm)
}

// SnapshotCreate creates VM snapshot
func SnapshotCreate(ctx context.Context, client *vim25.Client, params *jt.SnapshotCreateParams) (int32, error) {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return 0, err
	}

	beforeSnapshots, err := vmSnapshots(ctx, vm)
	if err != nil {
		return 0, err
	}

	task, err := vm.CreateSnapshot(ctx, params.Name, params.Description, params.Memory, params.Quiesce)
	if err != nil {
		return 0, err
	}

	if errWait := task.Wait(ctx); errWait != nil {
		return 0, errWait
	}

	afterSnapshots, err := vmSnapshots(ctx, vm)
	if err != nil {
		return 0, err
	}

	afterIDs := make([]int32, 0, len(afterSnapshots))
	for _, i := range afterSnapshots {
		afterIDs = append(afterIDs, i.ID)
	}

	beforeIDs := make([]int32, 0, len(beforeSnapshots))
	for _, i := range beforeSnapshots {
		beforeIDs = append(beforeIDs, i.ID)
	}

	// at the same time somebody can create another snapshot. So, also, check snapshot names. I hope it enough.
	for _, i := range diff(afterIDs, beforeIDs) {
		for _, s := range afterSnapshots {
			if s.ID == i && s.Name == params.Name {
				return s.ID, nil
			}
		}
	}

	return 0, errors.New("could not get snapshot ID")
}

// RestoreFromSnapshot restore VM from snapshot
func RestoreFromSnapshot(ctx context.Context, client *vim25.Client, params *jt.VMRestoreFromSnapshotParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	var o mo.VirtualMachine

	err = vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return err
	}

	if o.Snapshot == nil || len(o.Snapshot.RootSnapshotList) == 0 {
		return errors.New("no snapshots for this VM")
	}

	sRef := &snapshotReference{}
	sRef.findByID(o.Snapshot.RootSnapshotList, params.SnapshotID)
	if !sRef.exist {
		return fmt.Errorf("cound not find snapshot with id %d", params.SnapshotID)
	}

	task, err := vm.RevertToSnapshot(ctx, sRef.value(), params.PowerOn)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

// DeleteSnapshot deletes snapshot
func DeleteSnapshot(ctx context.Context, client *vim25.Client, params *jt.VMSnapshotDeleteParams) error {
	// snapshot lookup by name, where name can be:
	// 1) snapshot ManagedObjectReference.Value (unique)
	// 2) snapshot name (may not be unique)
	// 3) snapshot tree path (may not be unique)
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	var o mo.VirtualMachine

	err = vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return err
	}

	if o.Snapshot == nil || len(o.Snapshot.RootSnapshotList) == 0 {
		return errors.New("no snapshots for this VM")
	}

	sRef := &snapshotReference{}
	sRef.findByID(o.Snapshot.RootSnapshotList, params.SnapshotID)
	if !sRef.exist {
		return fmt.Errorf("cound not find snapshot with id %d", params.SnapshotID)
	}

	task, err := vm.RemoveSnapshot(ctx, sRef.value(), false, nil)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

func (s *snapshotReference) findByID(snapshots []types.VirtualMachineSnapshotTree, id int32) {
	for _, st := range snapshots {
		if id == st.Id {
			s.ref = st.Snapshot
			s.exist = true
		}
		s.findByID(st.ChildSnapshotList, id)
	}
}

func (s *snapshotReference) value() string {
	return s.ref.Value
}

func vmSnapshots(ctx context.Context, vm *object.VirtualMachine) ([]jt.Snapshot, error) {
	var o mo.VirtualMachine

	err := vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return nil, err
	}

	st := make([]jt.Snapshot, 0)
	if o.Snapshot == nil {
		return st, nil
	}

	ch := make(chan jt.Snapshot, 1000)
	walk(o.Snapshot.RootSnapshotList, ch)

	close(ch)
	for v := range ch {
		st = append(st, v)
	}

	return st, nil
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

func diff(slice1 []int32, slice2 []int32) []int32 {
	var diff []int32

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

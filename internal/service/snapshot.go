package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/vim25/mo"
	vmware_types "github.com/vmware/govmomi/vim25/types"

	"github.com/vterdunov/janna-api/internal/domain"
	"github.com/vterdunov/janna-api/internal/types"
)

func (s *service) VMSnapshotsList(ctx context.Context, params *types.VMSnapshotsListParams) ([]domain.Snapshot, error) {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	return vmSnapshots(ctx, vm)
}

func (s *service) VMSnapshotCreate(ctx context.Context, params *types.SnapshotCreateParams) (int32, error) {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
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

type snapshotReference struct {
	ref   vmware_types.ManagedObjectReference
	exist bool
}

func (s *snapshotReference) findByID(snapshots []vmware_types.VirtualMachineSnapshotTree, id int32) {
	for i := range snapshots {
		st := &snapshots[i]
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

func (s *service) VMRestoreFromSnapshot(ctx context.Context, params *types.VMRestoreFromSnapshotParams) error {
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
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

func (s *service) VMSnapshotDelete(ctx context.Context, params *types.VMSnapshotDeleteParams) error {
	// snapshot lookup by name, where name can be:
	// 1) snapshot ManagedObjectReference.Value (unique)
	// 2) snapshot name (may not be unique)
	// 3) snapshot tree path (may not be unique)
	vm, err := findByUUID(ctx, s.Client, params.Datacenter, params.UUID)
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

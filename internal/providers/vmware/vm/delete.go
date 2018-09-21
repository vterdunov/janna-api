package vm

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/internal/types"
)

func Delete(ctx context.Context, client *vim25.Client, params *jt.VMDeleteParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	state, psErr := vm.PowerState(ctx)
	if psErr != nil {
		return errors.Wrap(psErr, "could not get Virtual Machine power state")
	}

	if state != types.VirtualMachinePowerStatePoweredOff {
		task, pOffErr := vm.PowerOff(ctx)
		if pOffErr != nil {
			return errors.Wrap(pOffErr, "could not power off Virtual Machine before destroying")
		}

		if err = task.Wait(ctx); err != nil {
			return errors.Wrap(err, "could not power off Virtual Machine before destroying")
		}
	}

	destroyTask, err := vm.Destroy(ctx)
	if err != nil {
		return err
	}

	return destroyTask.Wait(ctx)
}

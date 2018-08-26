package vm

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/vim25"
	jt "github.com/vterdunov/janna-api/internal/types"
)

func Delete(ctx context.Context, client *vim25.Client, params *jt.VMDeleteParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	task, err := vm.PowerOff(ctx)
	if err != nil {
		return errors.Wrap(err, "could not power off Virtual Machine before destroying")
	}

	if err := task.Wait(ctx); err != nil {
		return errors.Wrap(err, "could not power off Virtual Machine before destroying")
	}

	task, err = vm.Destroy(ctx)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

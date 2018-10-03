package vm

import (
	"context"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/internal/types"
)

// Power changes VM power state
func Power(ctx context.Context, client *vim25.Client, params *jt.VMPowerParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	var task *object.Task

	switch params.State {
	case "on":
		task, err = vm.PowerOn(ctx)
	case "off":
		task, err = vm.PowerOff(ctx)
	case "reset":
		task, err = vm.Reset(ctx)
	case "suspend":
		task, err = vm.Suspend(ctx)
	case "reboot":
		err = vm.RebootGuest(ctx)

		if err != nil && params.Force && isToolsUnavailable(err) {
			task, err = vm.Reset(ctx)
		}
	case "shutdown":
		err = vm.ShutdownGuest(ctx)

		if err != nil && params.Force && isToolsUnavailable(err) {
			task, err = vm.PowerOff(ctx)
		}
	}

	if err != nil {
		return err
	}

	if task != nil {
		err = task.Wait(ctx)
	}

	if err == nil {
		return nil
	}

	return err

}

func isToolsUnavailable(err error) bool {
	if soap.IsSoapFault(err) {
		soapFault := soap.ToSoapFault(err)
		if _, ok := soapFault.VimFault().(types.ToolsUnavailable); ok {
			return ok
		}
	}

	return false
}

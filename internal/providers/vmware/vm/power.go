package vm

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"

	jt "github.com/vterdunov/janna-api/internal/types"
)

// Power changes VM power state
func Power(ctx context.Context, client *vim25.Client, params *jt.VMPowerParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}
	_ = vm

	var task *object.Task

	switch params.State {
	case "on":
		fmt.Println("Powering on")
		// task, err = vm.PowerOn(ctx)
	case "off":
		fmt.Println("Powering off")
		// task, err = vm.PowerOff(ctx)
	case "reser":
		fmt.Println("Reset")
		// task, err = vm.Reset(ctx)
	case "suspend":
		fmt.Println("Suspend")
		// task, err = vm.Suspend(ctx)
	case "reboot":
		fmt.Println("Reboot guest")
		// err = vm.RebootGuest(ctx)

		// if err != nil && cmd.Force && isToolsUnavailable(err) {
		// task, err = vm.Reset(ctx)
		// }
	case "shutdown":
		fmt.Println("Shutdown guest")
		// err = vm.ShutdownGuest(ctx)

		// if err != nil && cmd.Force && isToolsUnavailable(err) {
		// task, err = vm.PowerOff(ctx)
		// }
	}

	if err != nil {
		return err
	}

	if task != nil {
		err = task.Wait(ctx)
	}

	if err == nil {
		fmt.Println("OK")
		return nil
	}

	return err

}

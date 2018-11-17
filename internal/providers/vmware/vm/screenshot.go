package vm

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/vmware/govmomi/vim25/soap"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	jt "github.com/vterdunov/janna-api/internal/types"
)

func Screenshot(ctx context.Context, client *vim25.Client, params *jt.VMScreenshotParams) ([]byte, error) {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	state, psErr := vm.PowerState(ctx)
	if psErr != nil {
		return nil, errors.Wrap(psErr, "could not get Virtual Machine power state")
	}

	if state != types.VirtualMachinePowerStatePoweredOn {
		return nil, fmt.Errorf("vm is not powered on (%s)", state)
	}

	u := client.URL()
	u.Path = "/screen"
	query := url.Values{"id": []string{vm.Reference().Value}}
	u.RawQuery = query.Encode()

	param := soap.DefaultDownload

	rc, _, derr := client.Download(ctx, u, &param)
	if derr != nil {
		return nil, derr
	}
	defer rc.Close()

	s, rErr := ioutil.ReadAll(rc)
	if rErr != nil {
		return nil, rErr
	}

	return s, nil
}

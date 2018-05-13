package vm

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vterdunov/janna-api/pkg/config"
	jannatypes "github.com/vterdunov/janna-api/pkg/types"
)

// Info returns summary information about Virtual Machines
func Info(ctx context.Context, vmName string, logger log.Logger, cfg *config.Config, client *vim25.Client) (*jannatypes.VMSummary, error) {
	f := find.NewFinder(client, true)

	dc, err := f.DatacenterOrDefault(ctx, cfg.VMWare.DC)
	if err != nil {
		logger.Log("err", err)
		return nil, err
	}

	f.SetDatacenter(dc)

	vmObjs, err := f.VirtualMachineList(ctx, vmName)
	if err != nil {
		if _, ok := err.(*find.NotFoundError); ok {
			logger.Log("err", err)
			return nil, err
		}
		logger.Log("err", err)
		return nil, err
	}

	refs := make([]types.ManagedObjectReference, 0, len(vmObjs))
	for _, vm := range vmObjs {
		refs = append(refs, vm.Reference())
	}

	var vms []mo.VirtualMachine
	// Retrieve all properties
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var props []string
	props = nil

	pc := property.DefaultCollector(client)

	if len(refs) != 0 {
		err = pc.Retrieve(ctx, refs, props, &vms)
		if err != nil {
			logger.Log("err", err)
			return nil, err
		}
	}

	logger.Log(
		"count", len(vms),
		"msg", "Virtual machines found",
	)

	sum := &jannatypes.VMSummary{}
	for _, vmInfo := range vms {
		sum.Guest = vmInfo.Guest
		sum.Heartbeat = vmInfo.GuestHeartbeatStatus
		sum.Runtime = vmInfo.Summary.Runtime
		sum.Config = vmInfo.Summary.Config
	}
	return sum, nil
}

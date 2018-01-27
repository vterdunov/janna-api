package vm

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vterdunov/janna-api/config"
	jannatypes "github.com/vterdunov/janna-api/types"
)

// Info returns summary information about Virtual Machines
func Info(ctx context.Context, vmName string, logger log.Logger, cfg *config.Config) (jannatypes.VMSummary, error) {
	sum := jannatypes.VMSummary{}
	vmWareURL := cfg.Vmware.URL

	u, err := soap.ParseURL(vmWareURL)
	if err != nil {
		logger.Log("err", "cannot parse VMWare URL")
		return sum, err
	}

	insecure := cfg.Vmware.Insecure

	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		logger.Log("err", err)
		return sum, err
	}

	defer c.Logout(ctx)
	f := find.NewFinder(c.Client, true)

	dcName := cfg.Vmware.DC
	dc, err := f.DatacenterOrDefault(ctx, dcName)
	if err != nil {
		logger.Log("err", err)
		return sum, err
	}

	f.SetDatacenter(dc)

	vmObjs, err := f.VirtualMachineList(ctx, vmName)
	if err != nil {
		if _, ok := err.(*find.NotFoundError); ok {
			logger.Log("err", err)
			return sum, err
		}
		logger.Log("err", err)
		return sum, err
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

	pc := property.DefaultCollector(c.Client)

	if len(refs) != 0 {
		err = pc.Retrieve(ctx, refs, props, &vms)
		if err != nil {
			logger.Log("err", err)
			return sum, err
		}
	}

	logger.Log(
		"count", len(vms),
		"msg", "Virtual machines found",
	)

	for _, vmInfo := range vms {
		sum.Guest = vmInfo.Guest
		sum.Heartbeat = vmInfo.GuestHeartbeatStatus
		sum.Runtime = vmInfo.Summary.Runtime
		sum.Config = vmInfo.Summary.Config
	}
	return sum, nil
}

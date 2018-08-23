package vm

import (
	"context"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/internal/types"
)

// Info returns summary information about Virtual Machines
func Info(ctx context.Context, client *vim25.Client, params *jt.VMInfoParams) (*jt.VMSummary, error) {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	refs := make([]types.ManagedObjectReference, 0)
	refs = append(refs, vm.Reference())

	// Retrieve all properties
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var mVM mo.VirtualMachine
	var props []string

	pc := property.DefaultCollector(client)

	if err := pc.Retrieve(ctx, refs, props, &mVM); err != nil {
		return nil, err
	}

	sum := &jt.VMSummary{
		Guest:     mVM.Guest,
		Heartbeat: mVM.GuestHeartbeatStatus,
		Runtime:   mVM.Summary.Runtime,
		Config:    mVM.Summary.Config,
	}

	return sum, nil
}

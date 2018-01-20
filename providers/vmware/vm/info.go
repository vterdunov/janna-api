package vm

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	jannatypes "github.com/vterdunov/janna-api/types"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// Info returns summary information about Virtual Machines
func Info(ctx context.Context, vmName string) (jannatypes.VMSummary, error) {
	sum := jannatypes.VMSummary{}

	vmWareURI := os.Getenv("VMWARE_URI")
	u, _ := soap.ParseURL(vmWareURI)

	// TODO: Get from config
	insecure := true
	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		log.Error().Err(err).Msg("Cannot create govmomi client")
	}

	defer c.Logout(ctx)
	f := find.NewFinder(c.Client, true)

	dc, err := f.DatacenterOrDefault(ctx, "/DC1")
	if err != nil {
		log.Error().Err(err).Msg("Cannot found the Datacenter")
	}

	f.SetDatacenter(dc)

	vmObjs, err := f.VirtualMachineList(ctx, vmName)
	if err != nil {
		if _, ok := err.(*find.NotFoundError); ok {
			log.Error().Err(err).Msg("Cannot found VM")
			return sum, err
		}
		log.Error().Err(err)
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
			log.Error().Msg("Cannot retreive inforamtion about VM")
			return sum, err
		}
	}

	log.Info().Int("count", len(vms)).Msg("Virtual machines found")

	for _, vmInfo := range vms {
		sum.Guest = vmInfo.Guest
		sum.Heartbeat = vmInfo.GuestHeartbeatStatus
		sum.Runtime = vmInfo.Summary.Runtime
		sum.Config = vmInfo.Summary.Config
	}
	return sum, nil
}

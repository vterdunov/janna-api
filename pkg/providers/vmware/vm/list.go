package vm

import (
	"context"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"

	"github.com/go-kit/kit/log"
	"github.com/vmware/govmomi/vim25"

	jt "github.com/vterdunov/janna-api/pkg/types"
)

// List returns a list of VM names and its UUIDs
func List(ctx context.Context, c *vim25.Client, params *jt.VMListParams, logger log.Logger) ([]string, error) {
	// this implementation works two times faster on my workload that using finder. 280VMs gets around 2.3sec vs ~6.6.sec using finder
	// but I don't know how to get machines from specific Datacenter, not all ESXi/vSPhere hosts.
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all machines
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, vm := range vms {
		// fmt.Println(vm.Summary.)
		// fmt.Printf("%s: %s\n", vm.Summary.Config.Name, vm.Summary.Config.Uuid)
		res = append(res, vm.Summary.Config.Name)
	}

	// // Implementation using finder. Slow.
	// f := find.NewFinder(c, true)
	// dc, err := f.DatacenterOrDefault(ctx, params.Datacenter)
	// if err != nil {
	// 	return nil, err
	// }

	// f.SetDatacenter(dc)

	// vms, err := f.VirtualMachineList(ctx, "*")
	// if err != nil {
	// 	return nil, err
	// }

	// pc := property.DefaultCollector(c)

	// // Convert datastores into list of references
	// var refs []types.ManagedObjectReference
	// for _, vm := range vms {
	// 	refs = append(refs, vm.Reference())
	// }

	// var vmt []mo.VirtualMachine
	// if err = pc.Retrieve(ctx, refs, []string{"summary"}, &vmt); err != nil {
	// 	return nil, err
	// }

	// for _, vm := range vmt {
	// 	fmt.Printf("%s: %s\n", vm.Summary.Config.Name, vm.Summary.Config.Uuid)
	// }

	return res, nil
}

package vm

import (
	"context"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/internal/types"
)

// List returns a list of VM names and its UUIDs
func List(ctx context.Context, c *vim25.Client, params *jt.VMListParams) (map[string]string, error) {

	root, err := chooseRoot(ctx, c, params)
	if err != nil {
		return nil, err
	}

	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, root, []string{"VirtualMachine"}, true)
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

	res := make(map[string]string)
	for _, vm := range vms {
		res[vm.Summary.Config.Uuid] = vm.Summary.Config.Name
	}

	return res, nil
}

func chooseRoot(ctx context.Context, c *vim25.Client, params *jt.VMListParams) (types.ManagedObjectReference, error) {
	var ref types.ManagedObjectReference
	f := find.NewFinder(c, true)
	dc, err := f.DatacenterOrDefault(ctx, params.Datacenter)
	if err != nil {
		return ref, err
	}

	if params.Folder != "" {
		f.SetDatacenter(dc)
		rp, err := f.FolderOrDefault(ctx, params.Folder)
		if err != nil {
			return ref, err
		}
		return rp.Reference(), nil
	}

	if params.ResourcePool != "" {
		f.SetDatacenter(dc)
		rp, err := f.ResourcePoolOrDefault(ctx, params.ResourcePool)
		if err != nil {
			return ref, err
		}
		return rp.Reference(), nil
	}
	return dc.Reference(), nil
}

package vm

import (
	"context"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/pkg/types"
)

// Find search Virtual Machine from given Datacenter and Path
func Find(ctx context.Context, client *vim25.Client, params *jt.VMFindParams) (map[string]string, error) {
	oVM, err := FindByPath(ctx, client, params.Datacenter, params.Path)
	if err != nil {
		return nil, err
	}

	refs := []types.ManagedObjectReference{oVM.Reference()}

	var vm mo.VirtualMachine

	pc := property.DefaultCollector(client)

	if err := pc.Retrieve(ctx, refs, []string{"summary"}, &vm); err != nil {
		return nil, err
	}

	res := make(map[string]string)
	res[vm.Summary.Config.Uuid] = vm.Summary.Config.Name

	return res, nil
}

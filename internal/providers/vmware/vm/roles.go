package vm

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"

	jt "github.com/vterdunov/janna-api/internal/types"
)

// RolesList get VM roles
func RolesList(ctx context.Context, client *vim25.Client, params *jt.VMRolesListParams) ([]jt.Role, error) {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return nil, err
	}

	am := object.NewAuthorizationManager(client)

	perms, err := am.RetrieveEntityPermissions(ctx, vm.Reference(), true)
	if err != nil {
		return nil, err
	}

	for _, p := range perms {
		_ = p
		// fmt.Println(p.Principal)
	}

	roles, err := am.RoleList(ctx)
	if err != nil {
		return nil, err
	}

	rr := []jt.Role{}
	for _, role := range roles {
		desc := role.Info.GetDescription()
		r := jt.Role{
			Name: role.Name,
			ID:   role.RoleId,
		}

		r.Description.Label = desc.Label
		r.Description.Summary = desc.Summary
		rr = append(rr, r)
	}

	// TODO: Implement get role name from IDs
	return rr, errors.New("not Implemented")
}

// AddRole adds role to specific VM
func AddRole(ctx context.Context, client *vim25.Client, params *jt.VMAddRoleParams) error {
	vm, err := FindByUUID(ctx, client, params.Datacenter, params.UUID)
	if err != nil {
		return err
	}

	p := types.Permission{
		Principal: params.Principal,
		RoleId:    params.RoleID,
	}
	pp := []types.Permission{}
	pp = append(pp, p)

	am := object.NewAuthorizationManager(client)
	if err := am.SetEntityPermissions(ctx, vm.Reference(), pp); err != nil {
		return err
	}

	return nil
}

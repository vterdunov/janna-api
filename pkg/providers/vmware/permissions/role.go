package permissions

import (
	"context"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"

	jt "github.com/vterdunov/janna-api/pkg/types"
)

// RoleList get all roles
func RoleList(ctx context.Context, client *vim25.Client) ([]jt.Role, error) {
	am := object.NewAuthorizationManager(client)
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

	return rr, err
}

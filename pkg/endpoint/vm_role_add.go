// nolint: dupl
package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/types"
)

// MakeVMAddRoleEndpoint returns an endpoint via the passed service
func MakeVMAddRoleEndpoint(s service.Service) endpoint.Endpoint { // nolint: dupl
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMAddRoleRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMAddRoleParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
			Principal:  req.Principal,
			RoleID:     req.RoleID,
		}
		params.FillEmptyFields(s.GetConfig())

		err = s.VMAddRole(ctx, params)
		return VMAddRoleResponse{Err: err}, nil
	}
}

// VMAddRoleRequest collects the request parameters for the VMAddRole method
type VMAddRoleRequest struct {
	UUID       string
	Datacenter string `json:"datacenter"`
	Principal  string `json:"principal"`
	RoleID     int32  `json:"role_id"`
}

// VMAddRoleResponse collects the response values for the VMAddRole method
type VMAddRoleResponse struct {
	Err error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMAddRoleResponse) Failed() error {
	return r.Err
}

//nolint: dupl
package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMRolesListEndpoint returns an endpoint via the passed service
func MakeVMRolesListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMRolesListRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMRolesListParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		list, err := s.VMRolesList(ctx, params)
		return VMRolesListResponse{VMRolesList: list, Err: err}, nil
	}
}

// VMRolesListRequest collects the request parameters for the VMRolesList method
type VMRolesListRequest struct {
	UUID       string
	Datacenter string
}

// VMRolesListResponse collects the response values for the VMRolesList method
type VMRolesListResponse struct {
	VMRolesList []service.Role `json:"roles"`
	Err         error          `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMRolesListResponse) Failed() error {
	return r.Err
}

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
		if err != nil {
			return VMRolesListResponse{Err: err}, nil
		}

		rs := []VMRole{}
		for _, r := range list {
			role := VMRole{}
			role.Name = r.Name
			role.ID = r.ID
			role.Description.Label = r.Description.Label
			role.Description.Summary = r.Description.Summary
			rs = append(rs, role)
		}

		return VMRolesListResponse{VMRolesList: rs, Err: err}, nil
	}
}

// VMRolesListRequest collects the request parameters for the VMRolesList method
type VMRolesListRequest struct {
	UUID       string
	Datacenter string
}

// VMRolesListResponse collects the response values for the VMRolesList method
type VMRolesListResponse struct {
	VMRolesList []VMRole
	Err         error `json:"error,omitempty"`
}

type VMRole struct {
	Name        string `json:"name"`
	ID          int32  `json:"id"`
	Description struct {
		Label   string `json:"label"`
		Summary string `json:"summary"`
	} `json:"description"`
}

// Failed implements Failer
func (r VMRolesListResponse) Failed() error {
	return r.Err
}

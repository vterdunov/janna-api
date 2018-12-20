package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
)

// MakeRolesListEndpoint returns an endpoint via the passed service
func MakeRolesListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		roles, err := s.RoleList(ctx)
		rs := []Role{}

		for _, r := range roles {
			role := Role{}
			role.Name = r.Name
			role.ID = r.ID
			role.Description.Label = r.Description.Label
			role.Description.Summary = r.Description.Summary
			rs = append(rs, role)
		}
		return RoleListResponse{Roles: rs, Err: err}, nil
	}
}

// RoleListRequest collects the request parameters for the RoleList method
type RoleListRequest struct{}

// RoleListResponse collects the response values for the RoleList method
type RoleListResponse struct {
	Roles []Role
	Err   error `json:"error,omitempty"`
}

type Role struct {
	Name        string `json:"name"`
	ID          int32  `json:"id"`
	Description struct {
		Label   string `json:"label"`
		Summary string `json:"summary"`
	} `json:"description"`
}

// Failed implements Failer
func (r RoleListResponse) Failed() error {
	return r.Err
}

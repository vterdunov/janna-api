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
		return RoleListResponse{Roles: roles, Err: err}, nil
	}
}

// RoleListRequest collects the request parameters for the RoleList method
type RoleListRequest struct{}

// RoleListResponse collects the response values for the RoleList method
type RoleListResponse struct {
	Roles []service.Role `json:"roles"`
	Err   error          `json:"error,omitempty"`
}

// Failed implements Failer
func (r RoleListResponse) Failed() error {
	return r.Err
}

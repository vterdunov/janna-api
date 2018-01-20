package main

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/types"
)

// Endpoints collects all of the endpoints that compose the Service.
type Endpoints struct {
	VMInfoEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service. Useful in a profilesvc
// server.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		VMInfoEndpoint: MakeVMInfoEndpoint(s),
	}
}

// swagger:parameters
type vmInfoRequest struct {
	Name string `json:"name"`
}

// VM info data
// swagger:response
type vmInfoResponse struct {
	types.VMSummary
}

// MakeVMInfoEndpoint returns an endpoint via the passed service.
//
// swagger:route GET /vm/info vm vmInfo
//
// get information about VMs
//
// Responses:
//   200: vmInfoResponse
func MakeVMInfoEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(vmInfoRequest)
		v, _ := svc.VMInfo(ctx, req.Name)
		return vmInfoResponse{v}, nil
	}
}

package main

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/types"
)

// Endpoints collects all of the endpoints that compose the Service.
type Endpoints struct {
	InfoEndpoint    endpoint.Endpoint
	ReadyzEndpoint  endpoint.Endpoint
	HealthzEndpoint endpoint.Endpoint
	VMInfoEndpoint  endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		InfoEndpoint:    MakeInfoEndpoint(s),
		VMInfoEndpoint:  MakeVMInfoEndpoint(s),
		HealthzEndpoint: MakeHealthzEndpoint(s),
		ReadyzEndpoint:  MakeReadyzEndpoint(s),
	}
}

// MakeInfoEndpoint returns an endpoint via the passed service
//
// swagger:route GET /info app appInfo
//
// get information about the Service
//
// Responses:
//   200: InfoResponse
func MakeInfoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		b, c := s.Info()
		return InfoResponse{b, c}, nil
	}
}

// InfoResponse is the Service build information
// swagger:response
type InfoResponse struct {
	// in: body
	BuildTime string `json:"build_time"`
	// in: body
	Commit string `json:"commit"`
}

// MakeHealthzEndpoint returns an endpoint via the passed service
//
// swagger:route GET /healthz app appHealth
//
// liveness probe
//
// Responses:
//   200: healthzResponse
func MakeHealthzEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.Healthz()
		return readyzResponse{}, nil
	}
}

// Liveness probe
// swagger:response
type healthzResponse struct {
}

// MakeReadyzEndpoint returns an endpoint via the passed service
//
// swagger:route GET /readyz app appReady
//
// readiness probe
//
// Responses:
//   200: readyzResponse
func MakeReadyzEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.Readyz()
		return readyzResponse{}, nil
	}
}

// Readyness probe
// swagger:response
type readyzResponse struct {
}

// MakeVMInfoEndpoint returns an endpoint via the passed service
//
// swagger:route POST /vm/info vm vmInfo
//
// get information about VMs
//
// Responses:
//   200: vmInfoResponse
func MakeVMInfoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(vmInfoRequest)
		summary, err := s.VMInfo(ctx, req.Name)
		return vmInfoResponse{summary, err}, nil
	}
}

// swagger:parameters
type vmInfoRequest struct {
	Name string `json:"name"`
}

// VM info data
// swagger:response
type vmInfoResponse struct {
	Summary types.VMSummary `json:"summary,omitempty"`
	Err     error           `json:"error,omitempty"`
}

func (r vmInfoResponse) error() error {
	return r.Err
}

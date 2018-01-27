package main

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/types"
)

// Endpoints collects all of the endpoints that compose the Service.
type Endpoints struct {
	InfoEndpoint     endpoint.Endpoint
	ReadyzEndpoint   endpoint.Endpoint
	HealthzEndpoint  endpoint.Endpoint
	VMInfoEndpoint   endpoint.Endpoint
	VMDeployEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		InfoEndpoint:     MakeInfoEndpoint(s),
		HealthzEndpoint:  MakeHealthzEndpoint(s),
		ReadyzEndpoint:   MakeReadyzEndpoint(s),
		VMInfoEndpoint:   MakeVMInfoEndpoint(s),
		VMDeployEndpoint: MakeVMDeployEndpoint(s),
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
// swagger:route GET /vm vm vmInfo
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
	// in:body
	Summary types.VMSummary `json:"summary,omitempty"`
	// in:body
	Err error `json:"error,omitempty"`
}

func (r vmInfoResponse) error() error {
	return r.Err
}

// MakeVMDeployEndpoint returns an endpoint via the passed service
//
// swagger:route POST /vm vm vmInfo
//
// Create VM from OVA file
//
// Responses:
//   200: vmDeployResponse
func MakeVMDeployEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(vmDeployRequest)
		jid, err := s.VMDeploy(
			ctx,
			req.Name,
			req.OVAURL,
			req.Network,
			req.Datastores,
			req.Cluster,
			req.VMFolder,
		)

		return vmDeployResponse{jid, err}, nil
	}
}

// swagger:parameters
type vmDeployRequest struct {
	Name       string `json:"name"`
	OVAURL     string `json:"ova_url"`
	Network    string `json:"network,omitempty"`
	Datastores string `json:"datastores,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
	VMFolder   string `json:"vm_folder,omitempty"`
}

// VM deploy response fields
// swagger:response
type vmDeployResponse struct {
	// in:body
	JID int `json:"job_id,omitempty"`
	// in:body
	Err error `json:"error,omitempty"`
}

func (r vmDeployResponse) error() error {
	return r.Err
}

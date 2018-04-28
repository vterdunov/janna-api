package jannaendpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/vterdunov/janna-api/pkg/jannaservice"
	"github.com/vterdunov/janna-api/pkg/types"
)

// Endpoints collects all of the endpoints that compose the Service.
type Endpoints struct {
	InfoEndpoint     endpoint.Endpoint
	ReadyzEndpoint   endpoint.Endpoint
	HealthzEndpoint  endpoint.Endpoint
	VMInfoEndpoint   endpoint.Endpoint
	VMDeployEndpoint endpoint.Endpoint
}

// New returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func New(s jannaservice.Service, logger log.Logger) Endpoints {
	var infoEndpoint endpoint.Endpoint
	infoEndpoint = MakeInfoEndpoint(s)
	infoEndpoint = LoggingMiddleware(log.With(logger, "method", "Info"))(infoEndpoint)

	var healthzEndpoint endpoint.Endpoint
	healthzEndpoint = MakeHealthzEndpoint(s)
	healthzEndpoint = LoggingMiddleware(log.With(logger, "method", "Healthz"))(healthzEndpoint)

	var readyzEndpoint endpoint.Endpoint
	readyzEndpoint = MakeReadyzEndpoint(s)
	readyzEndpoint = LoggingMiddleware(log.With(logger, "method", "Readyz"))(readyzEndpoint)

	var vmInfoEndpoint endpoint.Endpoint
	vmInfoEndpoint = MakeVMInfoEndpoint(s)
	vmInfoEndpoint = LoggingMiddleware(log.With(logger, "method", "VMInfo"))(vmInfoEndpoint)

	var vmDeployEndpoint endpoint.Endpoint
	vmDeployEndpoint = MakeVMDeployEndpoint(s)
	vmDeployEndpoint = LoggingMiddleware(log.With(logger, "method", "VMDeploy"))(vmDeployEndpoint)

	return Endpoints{
		InfoEndpoint:     infoEndpoint,
		HealthzEndpoint:  healthzEndpoint,
		ReadyzEndpoint:   readyzEndpoint,
		VMInfoEndpoint:   vmInfoEndpoint,
		VMDeployEndpoint: vmDeployEndpoint,
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
func MakeInfoEndpoint(s jannaservice.Service) endpoint.Endpoint {
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
func MakeHealthzEndpoint(s jannaservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.Healthz()
		return healthzResponse{}, nil
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
func MakeReadyzEndpoint(s jannaservice.Service) endpoint.Endpoint {
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
func MakeVMInfoEndpoint(s jannaservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VMInfoRequest)
		summary, err := s.VMInfo(ctx, req.Name)
		return vmInfoResponse{summary, err}, nil
	}
}

// VMInfoRequest collects the request parameters for the VMInfo method.
// swagger:parameters
type VMInfoRequest struct {
	Name   string
	Folder string
}

// vmInfoResponse collects the response values for the VMInfo method
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
// swagger:route POST /vm vm vmDeploy
//
// Create VM from OVA file
//
// Responses:
//   200: vmDeployResponse
func MakeVMDeployEndpoint(s jannaservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(VMDeployRequest)
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
type VMDeployRequest struct {
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

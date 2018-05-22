package endpoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/types"
)

// Endpoints collects all of the endpoints that compose the Service.
type Endpoints struct {
	InfoEndpoint                  endpoint.Endpoint
	ReadyzEndpoint                endpoint.Endpoint
	HealthzEndpoint               endpoint.Endpoint
	VMListEndpoint                endpoint.Endpoint
	VMInfoEndpoint                endpoint.Endpoint
	VMDeployEndpoint              endpoint.Endpoint
	VMSnapshotsListEndpoint       endpoint.Endpoint
	VMSnapshotCreateEndpoint      endpoint.Endpoint
	VMRestoreFromSnapshotEndpoint endpoint.Endpoint
}

// New returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func New(s service.Service, logger log.Logger) Endpoints {
	infoEndpoint := MakeInfoEndpoint(s)
	infoEndpoint = LoggingMiddleware(log.With(logger, "method", "Info"))(infoEndpoint)

	healthzEndpoint := MakeHealthzEndpoint(s)

	readyzEndpoint := MakeReadyzEndpoint(s)

	vmListEndpoint := MakeVMListEndpoint(s)
	vmListEndpoint = LoggingMiddleware(log.With(logger, "method", "VMList"))(vmListEndpoint)

	vmInfoEndpoint := MakeVMInfoEndpoint(s)
	vmInfoEndpoint = LoggingMiddleware(log.With(logger, "method", "VMInfo"))(vmInfoEndpoint)

	vmDeployEndpoint := MakeVMDeployEndpoint(s, logger)
	vmDeployEndpoint = LoggingMiddleware(log.With(logger, "method", "VMDeploy"))(vmDeployEndpoint)

	vmSnapshotsListEndpoint := MakeVMSnapshotsListEndpoint(s)
	vmSnapshotsListEndpoint = LoggingMiddleware(log.With(logger, "method", "VMSnapshotsList"))(vmSnapshotsListEndpoint)

	vmSnapshotCreateEndpoint := MakeVMSnapshotCreateEndpoint(s)
	vmSnapshotCreateEndpoint = LoggingMiddleware(log.With(logger, "method", "VMSnapshotCreate"))(vmSnapshotCreateEndpoint)

	vmRestoreFromSnapshotEndpoint := MakeVMRestoreFromSnapshotEndpoint(s)
	vmRestoreFromSnapshotEndpoint = LoggingMiddleware(log.With(logger, "method", "VMRestoreFromSnapshot"))(vmRestoreFromSnapshotEndpoint)
	return Endpoints{
		InfoEndpoint:                  infoEndpoint,
		HealthzEndpoint:               healthzEndpoint,
		ReadyzEndpoint:                readyzEndpoint,
		VMListEndpoint:                vmListEndpoint,
		VMInfoEndpoint:                vmInfoEndpoint,
		VMDeployEndpoint:              vmDeployEndpoint,
		VMSnapshotsListEndpoint:       vmSnapshotsListEndpoint,
		VMSnapshotCreateEndpoint:      vmSnapshotCreateEndpoint,
		VMRestoreFromSnapshotEndpoint: vmRestoreFromSnapshotEndpoint,
	}
}

// Failer is an interface that should be implemented by response types.
// Response encoders can check if responses are Failer, and if so they've
// failed, and if so encode them using a separate write path based on the error.
type Failer interface {
	Failed() error
}

// MakeInfoEndpoint returns an endpoint via the passed service
func MakeInfoEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		b, c := s.Info()
		return InfoResponse{b, c}, nil
	}
}

// InfoResponse is the Service build information
type InfoResponse struct {
	BuildTime string `json:"build_time"`
	Commit    string `json:"commit"`
}

// MakeHealthzEndpoint returns an endpoint via the passed service
func MakeHealthzEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.Healthz()
		return healthzResponse{}, nil
	}
}

// Liveness probe
type healthzResponse struct{}

// MakeReadyzEndpoint returns an endpoint via the passed service
func MakeReadyzEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.Readyz()
		return readyzResponse{}, nil
	}
}

// Readyness probe
type readyzResponse struct{}

// MakeVMListEndpoint returns an endpoint via the passed service
func MakeVMListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VMListRequest)
		list, err := s.VMList(ctx, req.Folder)
		return VMListResponse{list, err}, nil
	}
}

// VMListRequest collects the request parameters for the VMList method
type VMListRequest struct {
	Folder string
}

// VMListResponse collects the response values for the VMList method
type VMListResponse struct {
	VMList []string `json:"vm_list,omitempty"`
	Err    error    `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMListResponse) Failed() error {
	return r.Err
}

// MakeVMInfoEndpoint returns an endpoint via the passed service
func MakeVMInfoEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VMInfoRequest)
		summary, err := s.VMInfo(ctx, req.Name)
		return VMInfoResponse{summary, err}, nil
	}
}

// VMInfoRequest collects the request parameters for the VMInfo method
type VMInfoRequest struct {
	Name   string
	Folder string
}

// VMInfoResponse collects the response values for the VMInfo method
type VMInfoResponse struct {
	Summary *types.VMSummary `json:"summary,omitempty"`
	Err     error            `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMInfoResponse) Failed() error {
	return r.Err
}

// MakeVMDeployEndpoint returns an endpoint via the passed service
func MakeVMDeployEndpoint(s service.Service, logger log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(VMDeployRequest)
		logger.Log("msg", "incoming request params", "params", fmt.Sprintf("%+v", req))

		// TODO: Try to write middleware that will validate parameters
		// Minimal validating incoming params
		if req.Name == "" || req.OVAURL == "" {
			return VMDeployResponse{0, errors.New("Invalid arguments. Pass reqired arguments")}, nil
		}

		params := &types.VMDeployParams{
			Name:       req.Name,
			OVAURL:     req.OVAURL,
			Datastores: req.Datastores,
			Networks:   req.Networks,
			Datacenter: req.Datacenter,
			Cluster:    req.Cluster,
			Folder:     req.Folder,
		}

		jid, err := s.VMDeploy(ctx, params)

		return VMDeployResponse{jid, err}, nil
	}
}

// VMDeployRequest collects the request parameters for the VMDeploy method
type VMDeployRequest struct {
	Name       string            `json:"name"`
	OVAURL     string            `json:"ova_url"`
	Datastores []string          `json:"datastores,omitempty"`
	Networks   map[string]string `json:"networks,omitempty"`
	Datacenter string            `json:"datacenter,omitempty"`
	Cluster    string            `json:"cluster,omitempty"`
	Folder     string            `json:"folder,omitempty"`
}

// VMDeployResponse fields
type VMDeployResponse struct {
	JID int   `json:"job_id,omitempty"`
	Err error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMDeployResponse) Failed() error {
	return r.Err
}

// MakeVMSnapshotsListEndpoint returns an endpoint via the passed service
func MakeVMSnapshotsListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VMSnapshotsListRequest)
		list, err := s.VMSnapshotsList(ctx, req.VMName)
		return VMSnapshotsListResponse{list, err}, nil
	}
}

// VMSnapshotsListRequest collects the request parameters for the VMSnapshotsList method
type VMSnapshotsListRequest struct {
	VMName string
}

// VMSnapshotsListResponse collects the response values for the VMSnapshotsList method
type VMSnapshotsListResponse struct {
	VMSnapshotsList []types.Snapshot `json:"snapshots"`
	Err             error            `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMSnapshotsListResponse) Failed() error {
	return r.Err
}

// MakeVMSnapshotCreateEndpoint creates VM snapshot
func MakeVMSnapshotCreateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VMSnapshotCreateRequest)

		params := &types.SnapshotCreateParams{
			VMName:      req.VMname,
			Name:        req.Name,
			Description: req.Description,
			Memory:      req.Memory,
			Quiesce:     req.Quiesce,
		}

		err = s.VMSnapshotCreate(ctx, params)
		return VMSnapshotCreateResponse{err}, nil
	}
}

// VMSnapshotCreateRequest collects the request parameters for the VMSnapshotCreate method
type VMSnapshotCreateRequest struct {
	VMname      string
	Name        string
	Description string
	Memory      bool
	Quiesce     bool
}

// VMSnapshotCreateResponse collects the response values for the VMSnapshotCreate method
type VMSnapshotCreateResponse struct {
	Err error `json:"error"`
}

// Failed implements Failer
func (r VMSnapshotCreateResponse) Failed() error {
	return r.Err
}

// MakeVMRestoreFromSnapshotEndpoint creates VM snapshot
func MakeVMRestoreFromSnapshotEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VMRestoreFromSnapshotRequest)

		params := &types.VMRestoreFromSnapshotParams{
			VMName:  req.VMname,
			Name:    req.Name,
			PowerOn: req.PowerOn,
		}

		err = s.VMRestoreFromSnapshot(ctx, params)
		return VMSRestoreFromSnapshotResponse{err}, nil
	}
}

// VMRestoreFromSnapshotRequest collects the request parameters for the VMRestoreFromSnapshot method
type VMRestoreFromSnapshotRequest struct {
	VMname  string
	Name    string
	PowerOn bool
}

// VMSRestoreFromSnapshotResponse collects the response values for the VMRestoreFromSnapshot method
type VMSRestoreFromSnapshotResponse struct {
	Err error `json:"error"`
}

// Failed implements Failer
func (r VMSRestoreFromSnapshotResponse) Failed() error {
	return r.Err
}

package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	_ "net/http/pprof" // Register pprof
	"strconv"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vterdunov/janna-api/internal/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
)

func populateRequestContext(ctx context.Context, r *http.Request) context.Context {
	ctx = context.WithValue(ctx, service.ContextKeyRequestXRequestID, r.Header.Get("X-Request-Id"))
	return ctx
}

// NewHTTPHandler mounts all of the service endpoints into an http.Handler.
func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger, debug bool) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerBefore(populateRequestContext),
	}

	r := mux.NewRouter()

	// Service state
	r.Path("/info").Methods("GET").Handler(httptransport.NewServer(
		endpoints.InfoEndpoint,
		decodeInfoRequest,
		encodeResponse,
		options...,
	))

	r.Path("/healthz").Methods("GET").Handler(httptransport.NewServer(
		endpoints.HealthzEndpoint,
		decodeHelthzRequest,
		encodeProbeResponse,
		options...,
	))

	r.Path("/readyz").Methods("GET").Handler(httptransport.NewServer(
		endpoints.ReadyzEndpoint,
		decodeReadyzRequest,
		encodeProbeResponse,
		options...,
	))

	r.Path("/metrics").Methods("GET").Handler(promhttp.Handler())

	if debug {
		r.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	}

	// Virtual Machines
	r.Path("/vms").Methods("GET").Handler(httptransport.NewServer(
		endpoints.VMListEndpoint,
		decodeVMListRequest,
		encodeVMListResponse,
		options...,
	))

	r.Path("/vms").Methods("POST").Handler(httptransport.NewServer(
		endpoints.VMDeployEndpoint,
		decodeVMDeployRequest,
		encodeResponse,
		options...,
	))

	r.Path("/vms/{vm}").Methods("GET").Handler(httptransport.NewServer(
		endpoints.VMInfoEndpoint,
		decodeVMInfoRequest,
		encodeResponse,
		options...,
	))

	r.Path("/vms/{vm}").Methods("DELETE").Handler(httptransport.NewServer(
		endpoints.VMDeleteEndpoint,
		decodeVMDeleteRequest,
		encodeResponse,
		options...,
	))

	// Snapshots
	r.Path("/vms/{vm}/snapshots").Methods("GET").Handler(httptransport.NewServer(
		endpoints.VMSnapshotsListEndpoint,
		decodeVMSnapshotsListyRequest,
		encodeResponse,
		options...,
	))

	r.Path("/vms/{vm}/snapshots").Methods("POST").Handler(httptransport.NewServer(
		endpoints.VMSnapshotCreateEndpoint,
		decodeVMSnapshotCreateRequest,
		encodeResponse,
		options...,
	))

	r.Path("/vms/{vm}/snapshots").Methods("DELETE").Handler(httptransport.NewServer(
		endpoints.VMSnapshotDeleteEndpoint,
		decodeVMSnapshotDeleteRequest,
		encodeResponse,
		options...,
	))

	r.Path("/vms/{vm}/revert/{snapshot}").Methods("POST").Handler(httptransport.NewServer(
		endpoints.VMRestoreFromSnapshotEndpoint,
		decodeVMRestoreFromSnapshotRequest,
		encodeResponse,
		options...,
	))

	// Power state
	r.Path("/vms/{vm}/power").Methods("PATCH").Handler(httptransport.NewServer(
		endpoints.VMPowerEndpoint,
		decodeVMPowerRequest,
		encodeResponse,
		options...,
	))

	// Read VM roles
	r.Path("/vms/{vm}/roles").Methods("GET").Handler(httptransport.NewServer(
		endpoints.VMRolesListEndpoint,
		decodeVMRolesListRequest,
		encodeVMRoleListResponse,
		options...,
	))

	// Add VM roles
	r.Path("/vms/{vm}/roles").Methods("PATCH").Handler(httptransport.NewServer(
		endpoints.VMAddRoleEndpoint,
		decodeVMAddRoleRequest,
		encodeResponse,
		options...,
	))

	// Get VM screenshot
	r.Path("/vms/{vm}/screenshot").Methods("GET").Handler(httptransport.NewServer(
		endpoints.VMScreenshotEndpoint,
		decodeVMScreenshotRequest,
		encodeVMScreenshotResponse,
		options...,
	))

	// Find VM
	r.Path("/find/vm").Methods("GET").Handler(httptransport.NewServer(
		endpoints.VMFindEndpoint,
		decodeVMFindRequest,
		encodeResponse,
		options...,
	))

	// Roles
	r.Path("/permissions/roles").Methods("GET").Handler(httptransport.NewServer(
		endpoints.RoleListEndpoint,
		decodeRoleListRequest,
		encodeRoleListResponse,
		options...,
	))

	// Tasks statuses
	r.Path("/tasks/{taskID}").Methods("GET").Handler(httptransport.NewServer(
		endpoints.TaskInfoEndpoint,
		decodeTaskInfoRequest,
		encodeTaskInfoResponse,
		options...,
	))

	r.Path("/openapi").Methods("GET").Handler(httptransport.NewServer(
		endpoints.OpenAPIEndpoint,
		decodeOpenAPIRequest,
		encodeOpenAPIResponse,
		options...,
	))

	return r
}

func decodeInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeHelthzRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeReadyzRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeVMListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMListRequest
	req.Folder = r.URL.Query().Get("folder")
	req.Datacenter = r.URL.Query().Get("datacenter")
	req.ResourcePool = r.URL.Query().Get("resource_pool")

	return req, nil
}

func decodeVMInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMInfoRequest
	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	req.Datacenter = r.URL.Query().Get("datacenter")

	return req, nil
}

func decodeVMDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMDeleteRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	err := json.NewDecoder(r.Body).Decode(&req)
	switch {
	case err == io.EOF:
		// Empty body. No operation.
	case err != nil:
		return nil, errors.Wrap(err, "Could not decode request: "+r.Method+" "+r.RequestURI)
	}

	return req, nil
}

func decodeVMFindRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMFindRequest

	req.Path = r.URL.Query().Get("path")
	req.Datacenter = r.URL.Query().Get("datacenter")

	return req, nil
}

func decodeVMDeployRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "Could not decode request: "+r.Method+" "+r.RequestURI)
	}
	return req, nil
}

func decodeVMSnapshotsListyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMSnapshotsListRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	req.Datacenter = r.URL.Query().Get("datacenter")

	return req, nil
}

func decodeVMSnapshotCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMSnapshotCreateRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "Could not decode request: "+r.Method+" "+r.RequestURI)
	}

	return req, nil
}

func decodeVMSnapshotDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMSnapshotDeleteRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "Could not decode request: "+r.Method+" "+r.RequestURI)
	}

	return req, nil
}

func decodeVMRestoreFromSnapshotRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMRestoreFromSnapshotRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	req.PowerOn = true

	sID, err := strconv.Atoi(vars["snapshot"])
	if err != nil {
		return nil, err
	}
	req.SnapshotID = int32(sID)

	return req, nil
}

func decodeVMPowerRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMPowerRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "Could not decode request: "+r.Method+" "+r.RequestURI)
	}

	return req, nil
}

func decodeVMRolesListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMRolesListRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	req.Datacenter = r.URL.Query().Get("datacenter")

	return req, nil
}

func decodeVMAddRoleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMAddRoleRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "Could not decode request: "+r.Method+" "+r.RequestURI)
	}

	return req, nil
}

func decodeVMScreenshotRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMScreenshotRequest

	vars := mux.Vars(r)
	req.UUID = vars["vm"]
	req.Datacenter = r.URL.Query().Get("datacenter")

	return req, nil
}

func decodeTaskInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.TaskInfoRequest

	vars := mux.Vars(r)
	req.TaskID = vars["taskID"]

	return req, nil
}

func decodeRoleListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.RoleListRequest
	return req, nil
}

func decodeOpenAPIRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

// common response decoder
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeBusinesLogicError(ctx, e.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeProbeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func encodeBusinesLogicError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeBusinesLogicError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func encodeVMRoleListResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeBusinesLogicError(ctx, e.Failed(), w)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp, ok := response.(endpoint.VMRolesListResponse)
	if !ok {
		encodeError(ctx, errors.New("could not parse VM summary"), w)
	}

	return json.NewEncoder(w).Encode(resp.VMRolesList)
}

func encodeRoleListResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeBusinesLogicError(ctx, e.Failed(), w)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp, ok := response.(endpoint.RoleListResponse)
	if !ok {
		encodeError(ctx, errors.New("could not parse VM summary"), w)
	}

	return json.NewEncoder(w).Encode(resp.Roles)
}

func encodeOpenAPIResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeBusinesLogicError(ctx, e.Failed(), w)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	res, ok := response.(endpoint.OpenAPIResponse)
	if !ok {
		encodeError(ctx, errors.New("could not get OpenAPI data"), w)
	}

	w.Write(res.Spec)
	return nil
}

func encodeVMScreenshotResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeBusinesLogicError(ctx, e.Failed(), w)
		return nil
	}

	w.Header().Set("Content-Type", "image/png")

	res, ok := response.(endpoint.VMScreenshotResponse)
	if !ok {
		encodeError(ctx, errors.New("could not get screenshot data"), w)
	}

	w.Write(res.Screen)
	return nil
}

func encodeTaskInfoResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeBusinesLogicError(ctx, e.Failed(), w)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	res, ok := response.(endpoint.TaskInfoResponse)
	if !ok {
		encodeError(ctx, errors.New("could not get OpenAPI data"), w)
	}
	return json.NewEncoder(w).Encode(res.Status)
}

func encodeVMListResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeBusinesLogicError(ctx, e.Failed(), w)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	res, ok := response.(endpoint.VMListResponse)
	if !ok {
		encodeError(ctx, errors.New("could not get Virtial Machines list"), w)
	}
	return json.NewEncoder(w).Encode(res.VMList)
}

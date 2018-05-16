package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/vterdunov/janna-api/pkg/endpoint"
)

// NewHTTPHandler mounts all of the service endpoints into an http.Handler.
func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
	}

	// Service state
	r.Methods("GET").Path("/info").Handler(httptransport.NewServer(
		endpoints.InfoEndpoint,
		decodeInfoRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/healthz").Handler(httptransport.NewServer(
		endpoints.HealthzEndpoint,
		decodeHelthzRequest,
		encodeProbeResponse,
		options...,
	))

	r.Methods("GET").Path("/readyz").Handler(httptransport.NewServer(
		endpoints.ReadyzEndpoint,
		decodeReadyzRequest,
		encodeProbeResponse,
		options...,
	))

	// Virtual Machines
	r.Methods("GET").Path("/vm").Handler(httptransport.NewServer(
		endpoints.VMListEndpoint,
		decodeVMListRequest,
		encodeNotImplemented,
		options...,
	))

	r.Methods("GET").Path("/vm/{vm}").Handler(httptransport.NewServer(
		endpoints.VMInfoEndpoint,
		decodeVMInfoRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/vm").Handler(httptransport.NewServer(
		endpoints.VMDeployEndpoint,
		decodeVMDeployRequest,
		encodeResponse,
		options...,
	))

	// Snapshots
	r.Methods("GET").Path("/vm/{vm}/snapshots").Handler(httptransport.NewServer(
		endpoints.VMSnapshotsListEndpoint,
		decodeVMSnapshotsListyRequest,
		encodeNotImplemented,
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

	return req, nil
}

func decodeVMInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMInfoRequest
	vars := mux.Vars(r)
	req.Name = vars["vm"]
	req.Folder = r.URL.Query().Get("folder")

	return req, nil
}

func decodeVMDeployRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "Could not decode request")
	}
	return req, nil
}

func decodeVMSnapshotsListyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.VMSnapshotsListRequest

	vars := mux.Vars(r)
	req.VMName = vars["vm"]

	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeProbeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func encodeNotImplemented(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	// check business logic errors
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
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

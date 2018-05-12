package jannatransport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/vterdunov/janna-api/pkg/jannaendpoint"
)

// NewHTTPHandler mounts all of the service endpoints into an http.Handler.
func NewHTTPHandler(endpoints jannaendpoint.Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
	}

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

	r.Methods("GET").Path("/vm").Handler(httptransport.NewServer(
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

func decodeVMInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req jannaendpoint.VMInfoRequest
	req.Name = r.URL.Query().Get("vmname")
	req.Folder = r.URL.Query().Get("folder")
	return req, nil
}

func decodeVMDeployRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req jannaendpoint.VMDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "Could not decode request")
	}
	return req, nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error.
type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeProbeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

// encode error
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

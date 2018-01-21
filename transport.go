package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// func My(ctx context.Context, code int, r *http.Request) {

// }

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	// f := My(context.Background(), 201, r)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		// httptransport.ServerFinalizer(f),
	}

	r.Methods("GET").Path("/info").Handler(httptransport.NewServer(
		e.InfoEndpoint,
		decodeInfoRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/healthz").Handler(httptransport.NewServer(
		e.HealthzEndpoint,
		decodeHelthzRequest,
		encodeProbeResponse,
		options...,
	))
	r.Methods("GET").Path("/readyz").Handler(httptransport.NewServer(
		e.ReadyzEndpoint,
		decodeReadyzRequest,
		encodeProbeResponse,
		options...,
	))

	r.Methods("POST").Path("/vm/info").Handler(httptransport.NewServer(
		e.VMInfoEndpoint,
		decodeVMInfoRequest,
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
	var req vmInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeProbeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

package main

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	// "log"
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
func MakeHTTPHandler(s Service) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	// options := []httptransport.ServerOption{
	// 	httptransport.ServerErrorLogger(logger),
	// 	httptransport.ServerErrorEncoder(encodeError),
	// }

	r.Methods("POST").Path("/vm/info").Handler(httptransport.NewServer(
		e.VMInfoEndpoint,
		decodeVMInfoRequest,
		encodeResponse,
	))

	return r
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

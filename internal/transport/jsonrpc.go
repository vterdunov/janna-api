package transport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http/jsonrpc"

	"github.com/vterdunov/janna-api/internal/endpoint"
)

// NewJSONRPCHandler returns a JSON RPC Server/Handler that can be passed to http.Handle()
func NewJSONRPCHandler(endpoints endpoint.Endpoints, logger log.Logger) *jsonrpc.Server {
	handler := jsonrpc.NewServer(
		makeEndpointCodecMap(endpoints),
		jsonrpc.ServerErrorLogger(logger),
	)
	return handler
}

// makeEndpointCodecMap returns a codec map configured for the service
func makeEndpointCodecMap(endpoints endpoint.Endpoints) jsonrpc.EndpointCodecMap {
	return jsonrpc.EndpointCodecMap{
		"vm_info": jsonrpc.EndpointCodec{
			Endpoint: endpoints.VMInfoEndpoint,
			Decode:   decodeJSONPRCVMInfoRequest,
			Encode:   encodeJSONRPCVMInfoResponse,
		},
		"vm_deploy": jsonrpc.EndpointCodec{
			Endpoint: endpoints.VMDeployEndpoint,
			Decode:   decodeJSONPRCVMDeployRequest,
			Encode:   encodeJSONRPCVMDeployResponse,
		},
	}
}

// VM Info
func decodeJSONPRCVMInfoRequest(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var req endpoint.VMInfoRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to sum request: %s", err),
		}
	}
	return req, nil
}

func encodeJSONRPCVMInfoResponse(_ context.Context, obj interface{}) (json.RawMessage, error) { //nolint: dupl
	res, ok := obj.(endpoint.VMInfoResponse)
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("Asserting result to *VMInfoResponse failed. Got %T, %+v", obj, obj),
		}
	}

	// check business logic errors
	if f, ok := obj.(endpoint.Failer); ok && f.Failed() != nil {
		return json.Marshal(f.Failed().Error())
	}

	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("could not marshal response: %s", err)
	}

	return b, nil
}

// VM Deploy
func decodeJSONPRCVMDeployRequest(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var req endpoint.VMDeployRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to sum request: %s", err),
		}
	}
	return req, nil
}

func encodeJSONRPCVMDeployResponse(_ context.Context, obj interface{}) (json.RawMessage, error) { //nolint: dupl
	res, ok := obj.(endpoint.VMDeployResponse)
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("Asserting result to *VMDeployResponse failed. Got %T, %+v", obj, obj),
		}
	}

	// check business logic errors
	if f, ok := obj.(endpoint.Failer); ok && f.Failed() != nil {
		return json.Marshal(f.Failed().Error())
	}

	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("could not marshal response: %s", err)
	}
	return b, nil
}

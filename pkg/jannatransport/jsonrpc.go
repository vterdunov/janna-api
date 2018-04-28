package jannatransport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/examples/addsvc/pkg/addendpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http/jsonrpc"
	"github.com/vterdunov/janna-api/pkg/jannaendpoint"
)

// NewJSONRPCHandler returns a JSON RPC Server/Handler that can be passed to http.Handle()
func NewJSONRPCHandler(endpoints jannaendpoint.Endpoints, logger log.Logger) *jsonrpc.Server {
	handler := jsonrpc.NewServer(
		makeEndpointCodecMap(endpoints),
		jsonrpc.ServerErrorLogger(logger),
	)
	return handler
}

// makeEndpointCodecMap returns a codec map configured for the jannaservice
func makeEndpointCodecMap(endpoints jannaendpoint.Endpoints) jsonrpc.EndpointCodecMap {
	return jsonrpc.EndpointCodecMap{
		"info": jsonrpc.EndpointCodec{
			Endpoint: endpoints.VMInfoEndpoint,
			Decode:   decodeJSONPRCVMInfoRequest,
			Encode:   encodeJSONRPCVMInfoResponse,
		},
	}
}

func decodeJSONPRCVMInfoRequest(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var req jannaendpoint.VMInfoRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to sum request: %s", err),
		}
	}
	return req, nil
}

func encodeJSONRPCVMInfoResponse(_ context.Context, obj interface{}) (json.RawMessage, error) {
	res, ok := obj.(addendpoint.SumResponse)
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("Asserting result to *SumResponse failed. Got %T, %+v", obj, obj),
		}
	}
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal response: %s", err)
	}
	return b, nil
}

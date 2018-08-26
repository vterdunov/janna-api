package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMDeleteEndpoint deletes a Virtual Machine
func MakeVMDeleteEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMDeleteRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMDeleteParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		err = s.VMDelete(ctx, params)
		return VMDeleteResponse{Err: err}, nil
	}
}

// VMDeleteRequest collects the request parameters for the VMDelete method
type VMDeleteRequest struct {
	UUID       string
	Datacenter string `json:"datacenter,omitempty"`
}

// VMDeleteResponse collects the response values for the VMDelete method
type VMDeleteResponse struct {
	Err error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMDeleteResponse) Failed() error {
	return r.Err
}

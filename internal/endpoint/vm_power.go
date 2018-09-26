package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMPowerEndpoint changes VM power state
func MakeVMPowerEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMPowerRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMPowerParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
			State:      req.State,
		}
		params.FillEmptyFields(s.GetConfig())

		err = s.VMPower(ctx, params)
		return VMPowerResponse{Err: err}, nil
	}
}

// VMPowerRequest collects the request parameters for the VMPower method
type VMPowerRequest struct {
	UUID       string
	Datacenter string `json:"datacenter"`
	State      string `json:"state"`
}

// VMPowerResponse collects the response values for the VMPower method
type VMPowerResponse struct {
	Err error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMPowerResponse) Failed() error {
	return r.Err
}

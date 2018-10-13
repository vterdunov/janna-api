package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMInfoEndpoint returns an endpoint via the passed service
func MakeVMInfoEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMInfoRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMInfoParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		summary, err := s.VMInfo(ctx, params)
		return VMInfoResponse{Summary: summary, Err: err}, nil
	}
}

// VMInfoRequest collects the request parameters for the VMInfo method
type VMInfoRequest struct {
	UUID       string
	Datacenter string
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

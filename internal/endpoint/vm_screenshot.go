package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMScreenshotEndpoint returns an endpoint via the passed service
func MakeVMScreenshotEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMScreenshotRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMScreenshotParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		s, err := s.VMScreenshot(ctx, params)
		return VMScreenshotResponse{Screen: s, Err: err}, nil
	}
}

// VMScreenshotRequest collects the request parameters for the VMScreenshot method
type VMScreenshotRequest struct {
	UUID       string
	Datacenter string
}

// VMScreenshotResponse collects the response values for the VMScreenshot method
type VMScreenshotResponse struct {
	Screen []byte
	Err    error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMScreenshotResponse) Failed() error {
	return r.Err
}

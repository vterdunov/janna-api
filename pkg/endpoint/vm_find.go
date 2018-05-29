package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/types"
)

// MakeVMFindEndpoint returns an endpoint via the passed service
func MakeVMFindEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMFindRequest)
		if !ok {
			return nil, errors.New("Could not parse request")
		}

		params := &types.VMFindParams{
			Path:       req.Path,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		vm, err := s.VMFind(ctx, params)
		return VMFindResponse{
			UUID: vm.UUID,
			Name: vm.Name,
			Err:  err,
		}, nil
	}
}

// VMFindRequest collects the request parameters for the VMFind method
type VMFindRequest struct {
	Path       string
	Datacenter string
}

// VMFindResponse collects the response values for the VMFind method
type VMFindResponse struct {
	UUID string `json:"uuid,omitempty"`
	Name string `json:"name,omitempty"`
	Err  error  `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMFindResponse) Failed() error {
	return r.Err
}

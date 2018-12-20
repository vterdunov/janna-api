package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMFindEndpoint returns an endpoint via the passed service
func MakeVMFindEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMFindRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMFindParams{
			Path:       req.Path,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		vm, err := s.VMFind(ctx, params)
		if err != nil {
			return VMFindResponse{Err: err}, nil
		}

		vmUUID := VMUuid{
			Name: vm.Name,
			UUID: vm.UUID,
		}

		return VMFindResponse{
			VMUuid: vmUUID,
			Err:    err,
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
	VMUuid
	Err error `json:"error,omitempty"`
}

type VMUuid struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

// Failed implements Failer
func (r VMFindResponse) Failed() error {
	return r.Err
}

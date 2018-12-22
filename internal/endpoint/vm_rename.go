//nolint: dupl
package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMRenameEndpoint changes VM power state
func MakeVMRenameEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMRenameRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMRenameParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
			Name:       req.Name,
		}
		params.FillEmptyFields(s.GetConfig())

		err = s.VMRename(ctx, params)
		return VMRenameResponse{Err: err}, nil
	}
}

// VMRenameRequest collects the request parameters for the VMRename method
type VMRenameRequest struct {
	UUID       string
	Datacenter string `json:"datacenter"`
	Name       string `json:"name"`
}

// VMRenameResponse collects the response values for the VMRename method
type VMRenameResponse struct {
	Err error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMRenameResponse) Failed() error {
	return r.Err
}

package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/types"
)

// MakeVMSnapshotDeleteEndpoint returns an endpoint via the passed service
func MakeVMSnapshotDeleteEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMSnapshotDeleteRequest)
		if !ok {
			return nil, errors.New("Could not parse request")
		}

		params := &types.VMSnapshotDeleteParams{
			UUID:       req.UUID,
			SnapshotID: req.SnapshotID,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		err = s.VMSnapshotDelete(ctx, params)
		return VMSnapshotDeleteResponse{err}, nil
	}
}

// VMSnapshotDeleteRequest collects the request parameters for the VMSnapshotDelete method
type VMSnapshotDeleteRequest struct {
	UUID       string
	SnapshotID int32  `json:"snapshot_id"`
	Datacenter string `json:"datacenter,omitempty"`
}

// VMSnapshotDeleteResponse collects the response values for the VMSnapshotDelete method
type VMSnapshotDeleteResponse struct {
	Err error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMSnapshotDeleteResponse) Failed() error {
	return r.Err
}

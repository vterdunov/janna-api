package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/types"
)

// MakeVMSnapshotsListEndpoint returns an endpoint via the passed service
func MakeVMSnapshotsListEndpoint(s service.Service) endpoint.Endpoint { // nolint: dupl
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMSnapshotsListRequest)
		if !ok {
			return nil, errors.New("Could not parse request")
		}

		params := &types.VMSnapshotsListParams{
			UUID:       req.UUID,
			Datacenter: req.Datacenter,
		}
		params.FillEmptyFields(s.GetConfig())

		list, err := s.VMSnapshotsList(ctx, params)
		return VMSnapshotsListResponse{list, err}, nil
	}
}

// VMSnapshotsListRequest collects the request parameters for the VMSnapshotsList method
type VMSnapshotsListRequest struct {
	UUID       string
	Datacenter string
}

// VMSnapshotsListResponse collects the response values for the VMSnapshotsList method
type VMSnapshotsListResponse struct {
	VMSnapshotsList []types.Snapshot `json:"snapshots"`
	Err             error            `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMSnapshotsListResponse) Failed() error {
	return r.Err
}
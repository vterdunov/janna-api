// nolint: dupl
package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/types"
)

// MakeVMRestoreFromSnapshotEndpoint creates VM snapshot
func MakeVMRestoreFromSnapshotEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMRestoreFromSnapshotRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMRestoreFromSnapshotParams{
			UUID:       req.UUID,
			SnapshotID: req.SnapshotID,
			Datacenter: req.Datacenter,
			PowerOn:    req.PowerOn,
		}
		params.FillEmptyFields(s.GetConfig())

		err = s.VMRestoreFromSnapshot(ctx, params)
		return VMSRestoreFromSnapshotResponse{Err: err}, nil
	}
}

// VMRestoreFromSnapshotRequest collects the request parameters for the VMRestoreFromSnapshot method
type VMRestoreFromSnapshotRequest struct {
	UUID       string
	Datacenter string `json:"datacenter"`
	SnapshotID int32  `json:"snapshot_id"`
	PowerOn    bool   `json:"power_on"`
}

// VMSRestoreFromSnapshotResponse collects the response values for the VMRestoreFromSnapshot method
type VMSRestoreFromSnapshotResponse struct {
	Err error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMSRestoreFromSnapshotResponse) Failed() error {
	return r.Err
}

package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMSnapshotCreateEndpoint creates VM snapshot
func MakeVMSnapshotCreateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMSnapshotCreateRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.SnapshotCreateParams{
			UUID:        req.UUID,
			Datacenter:  req.Datacenter,
			Name:        req.Name,
			Description: req.Description,
			Memory:      req.Memory,
			Quiesce:     req.Quiesce,
		}
		params.FillEmptyFields(s.GetConfig())

		id, err := s.VMSnapshotCreate(ctx, params)
		return VMSnapshotCreateResponse{SnapshotID: id, Err: err}, nil
	}
}

// VMSnapshotCreateRequest collects the request parameters for the VMSnapshotCreate method
type VMSnapshotCreateRequest struct {
	UUID        string
	Datacenter  string `json:"datacenter"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Memory      bool   `json:"memory"`
	Quiesce     bool   `json:"quiesce"`
}

// VMSnapshotCreateResponse collects the response values for the VMSnapshotCreate method
type VMSnapshotCreateResponse struct {
	SnapshotID int32 `json:"snapshot_id,omitempty"`
	Err        error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMSnapshotCreateResponse) Failed() error {
	return r.Err
}

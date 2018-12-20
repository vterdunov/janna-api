package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMListEndpoint returns an endpoint via the passed service
func MakeVMListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(VMListRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		params := &types.VMListParams{
			Datacenter:   req.Datacenter,
			Folder:       req.Folder,
			ResourcePool: req.ResourcePool,
		}
		params.FillEmptyFields(s.GetConfig())

		list, err := s.VMList(ctx, params)
		resVMuuid := []VMUuid{}
		for _, i := range list {
			u := VMUuid{
				Name: i.Name,
				UUID: i.UUID,
			}
			resVMuuid = append(resVMuuid, u)
		}

		return VMListResponse{VMList: resVMuuid, Err: err}, nil
	}
}

// VMListRequest collects the request parameters for the VMList method
type VMListRequest struct {
	Datacenter   string
	Folder       string
	ResourcePool string
}

// VMListResponse collects the response values for the VMList method
type VMListResponse struct {
	VMList []VMUuid
	Err    error `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMListResponse) Failed() error {
	return r.Err
}

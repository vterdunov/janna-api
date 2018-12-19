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
		respSummary := VMSummary{
			Name:             summary.Name,
			UUID:             summary.UUID,
			Template:         summary.Template,
			GuestID:          summary.GuestID,
			Annotation:       summary.Annotation,
			NumCpu:           summary.NumCpu,
			NumEthernetCards: summary.NumEthernetCards,
			NumVirtualDisks:  summary.NumVirtualDisks,
		}

		return VMInfoResponse{Summary: respSummary, Err: err}, nil
	}
}

// VMInfoRequest collects the request parameters for the VMInfo method
type VMInfoRequest struct {
	UUID       string
	Datacenter string
}

// VMInfoResponse collects the response values for the VMInfo method
type VMInfoResponse struct {
	Summary VMSummary
	Err     error `json:"error,omitempty"`
}

type VMSummary struct {
	Name             string `json:"name"`
	UUID             string `json:"uuid"`
	Template         bool   `json:"template"`
	GuestID          string `json:"guest_id"`
	Annotation       string `json:"annotation"`
	NumCpu           int32  `json:"num_cpu"`
	NumEthernetCards int32  `json:"num_ethernet_cards"`
	NumVirtualDisks  int32  `json:"num_virtual_disks"`
}

// Failed implements Failer
func (r VMInfoResponse) Failed() error {
	return r.Err
}

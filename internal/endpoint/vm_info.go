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
		if err != nil {
			return VMInfoResponse{Err: err}, nil
		}

		gi := VMGuestInfo{
			GuestID:            summary.VMGuestInfo.GuestID,
			GuestFullName:      summary.VMGuestInfo.GuestFullName,
			ToolsRunningStatus: summary.VMGuestInfo.ToolsRunningStatus,
			HostName:           summary.VMGuestInfo.HostName,
			IPAddress:          summary.VMGuestInfo.IPAddress,
		}

		return VMInfoResponse{
			Name:             summary.Name,
			UUID:             summary.UUID,
			Template:         summary.Template,
			GuestID:          summary.GuestID,
			Annotation:       summary.Annotation,
			PowerState:       summary.PowerState,
			NumCPU:           summary.NumCPU,
			NumEthernetCards: summary.NumEthernetCards,
			NumVirtualDisks:  summary.NumVirtualDisks,
			VMGuestInfo:      gi,
			Err:              err,
		}, nil
	}
}

// VMInfoRequest collects the request parameters for the VMInfo method
type VMInfoRequest struct {
	UUID       string
	Datacenter string
}

// VMInfoResponse collects the response values for the VMInfo method
type VMInfoResponse struct {
	Name             string `json:"name"`
	UUID             string `json:"uuid"`
	GuestID          string `json:"guest_id"`
	Annotation       string `json:"annotation"`
	PowerState       string `json:"power_state"`
	NumCPU           int32  `json:"num_cpu"`
	NumEthernetCards int32  `json:"num_ethernet_cards"`
	NumVirtualDisks  int32  `json:"num_virtual_disks"`
	Template         bool   `json:"template"`
	VMGuestInfo      `json:"guest_info"`
	Err              error `json:"error,omitempty"`
}

type VMGuestInfo struct {
	GuestID            string `json:"guest_id"`
	GuestFullName      string `json:"guest_full_name"`
	ToolsRunningStatus string `json:"tools_running_status"`
	HostName           string `json:"host_name"`
	IPAddress          string `json:"ip_address"`
}

// Failed implements Failer
func (r VMInfoResponse) Failed() error {
	return r.Err
}

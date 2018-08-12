package endpoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/vterdunov/janna-api/pkg/service"
	"github.com/vterdunov/janna-api/pkg/types"
)

// MakeVMDeployEndpoint returns an endpoint via the passed service
func MakeVMDeployEndpoint(s service.Service, logger log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(VMDeployRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		logger.Log("msg", "incoming request params", "params", fmt.Sprintf("%+v", req))

		// TODO: Try to write middleware that will validate parameters
		// Minimal validating incoming params
		if req.Name == "" || req.OVAURL == "" {
			return VMDeployResponse{JID: "", Err: errors.New("invalid arguments. Pass reqired arguments")}, nil
		}

		params := &types.VMDeployParams{
			Name:       req.Name,
			OVAURL:     req.OVAURL,
			Datastores: req.Datastores,
			Networks:   req.Networks,
			Datacenter: req.Datacenter,
			Cluster:    req.Cluster,
			Folder:     req.Folder,
		}

		jid, err := s.VMDeploy(ctx, params)

		return VMDeployResponse{JID: jid, Err: err}, nil
	}
}

// VMDeployRequest collects the request parameters for the VMDeploy method
type VMDeployRequest struct {
	Name       string            `json:"name"`
	OVAURL     string            `json:"ova_url"`
	Datastores []string          `json:"datastores,omitempty"`
	Networks   map[string]string `json:"networks,omitempty"`
	Datacenter string            `json:"datacenter,omitempty"`
	Cluster    string            `json:"cluster,omitempty"`
	Folder     string            `json:"folder,omitempty"`
}

// VMDeployResponse fields
type VMDeployResponse struct {
	JID string `json:"task_id,omitempty"`
	Err error  `json:"error,omitempty"`
}

// Failed implements Failer
func (r VMDeployResponse) Failed() error {
	return r.Err
}

package endpoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/internal/types"
)

// MakeVMDeployEndpoint returns an endpoint via the passed service
func MakeVMDeployEndpoint(s service.Service, logger log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(VMDeployRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		logger.Log("msg", "incoming request params", "params", req.String())

		// TODO: Try to write middleware that will validate parameters
		// Minimal validating incoming params
		if req.Name == "" || req.OVAURL == "" {
			return VMDeployResponse{JID: "", Err: errors.New("invalid arguments. Pass reqired arguments")}, nil
		}

		params := &types.VMDeployParams{
			Name:         req.Name,
			OVAURL:       req.OVAURL,
			Datastores:   req.Datastores,
			Networks:     req.Networks,
			Datacenter:   req.Datacenter,
			Host:         req.Host,
			ResourcePool: req.ResourcePool,
			Folder:       req.Folder,
			Annotation:   req.Annotation,
		}

		params.FillEmptyFields(s.GetConfig())

		jid, err := s.VMDeploy(ctx, params)

		return VMDeployResponse{JID: jid, Err: err}, nil
	}
}

// VMDeployRequest collects the request parameters for the VMDeploy method
type VMDeployRequest struct {
	Name         string            `json:"name"`
	OVAURL       string            `json:"ova_url"`
	Datastores   []string          `json:"datastores,omitempty"`
	Networks     map[string]string `json:"networks,omitempty"`
	Datacenter   string            `json:"datacenter,omitempty"`
	Host         string            `json:"host,omitempty"`
	ResourcePool string            `json:"resource_pool,omitempty"`
	Folder       string            `json:"folder,omitempty"`
	Annotation   string            `json:"annotation"`
}

func (r *VMDeployRequest) String() string {
	return fmt.Sprintf("name: %s, ova_url: %s, datastores: %s, networks: %s, datacenter: %s, host: %s, resource_pool: %s, folder: %s, annotation: %s",
		r.Name, r.OVAURL, r.Datastores, r.Networks, r.Datacenter, r.Host, r.ResourcePool, r.Folder, r.Annotation)
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

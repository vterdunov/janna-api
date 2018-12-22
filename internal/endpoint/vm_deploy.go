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

		// Minimal validating incoming params
		if req.Name == "" || req.OVAURL == "" {
			return VMDeployResponse{JID: "", Err: errors.New("invalid arguments. Pass reqired arguments")}, nil
		}

		params := &types.VMDeployParams{
			Name:       req.Name,
			OVAURL:     req.OVAURL,
			Datacenter: req.Datacenter,
			Folder:     req.Folder,
			Annotation: req.Annotation,
			Networks:   req.Networks,
			Datastores: struct {
				Type  string
				Names []string
			}{
				Type:  req.Datastores.Type,
				Names: req.Datastores.Names,
			},
			ComputerResources: struct {
				Path string
				Type string
			}{
				Path: req.ComputerResources.Path,
				Type: req.ComputerResources.Type,
			},
		}

		params.FillEmptyFields(s.GetConfig())

		jid, err := s.VMDeploy(ctx, params)

		return VMDeployResponse{JID: jid, Err: err}, nil
	}
}

// VMDeployRequest collects the request parameters for the VMDeploy method
type VMDeployRequest struct {
	Name              string            `json:"name"`
	OVAURL            string            `json:"ova_url"`
	Datacenter        string            `json:"datacenter,omitempty"`
	Folder            string            `json:"folder,omitempty"`
	Annotation        string            `json:"annotation"`
	Networks          map[string]string `json:"networks,omitempty"`
	Datastores        `json:"datastores"`
	ComputerResources `json:"computer_resources"`
}

type Datastores struct {
	Type  string
	Names []string
}

type ComputerResources struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

func (r *VMDeployRequest) String() string {
	return fmt.Sprintf("name: %s, ova_url: %s, datastores: %s, networks: %s, datacenter: %s, computer_resources: %s, folder: %s, annotation: %s",
		r.Name, r.OVAURL, r.Datastores, r.Networks, r.Datacenter, r.ComputerResources, r.Folder, r.Annotation)
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

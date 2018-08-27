package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
)

// MakeOpenAPIEndpoint returns an endpoint via the passed service
func MakeOpenAPIEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		spec, err := s.OpenAPI()
		if err != nil {
			return OpenAPIResponse{nil}, err
		}

		return OpenAPIResponse{Spec: spec}, nil
	}
}

// OpenAPIResponse is the Service build information
type OpenAPIResponse struct {
	Spec []byte
}

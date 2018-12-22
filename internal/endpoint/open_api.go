package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/vterdunov/janna-api/internal/service"
)

// MakeOpenAPIEndpoint returns an endpoint via the passed service
func MakeOpenAPIEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		spec, err := s.OpenAPI(ctx)
		if err != nil {
			return OpenAPIResponse{Spec: nil, Err: err}, nil
		}

		return OpenAPIResponse{Spec: spec}, nil
	}
}

// OpenAPIResponse serves OpenAPI specification
type OpenAPIResponse struct {
	Spec []byte
	Err  error `json:"error,omitempty"`
}

// Failed implements Failer
func (r OpenAPIResponse) Failed() error {
	return r.Err
}

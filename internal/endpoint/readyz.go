package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
)

// MakeReadyzEndpoint returns an endpoint via the passed service
func MakeReadyzEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.Readyz()
		return readyzResponse{}, nil
	}
}

// Readyness probe
type readyzResponse struct{}

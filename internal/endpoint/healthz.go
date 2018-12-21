package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/vterdunov/janna-api/internal/service"
)

// MakeHealthzEndpoint returns an endpoint via the passed service
func MakeHealthzEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		s.Healthz()
		return healthzResponse{}, nil
	}
}

// Liveness probe
type healthzResponse struct{}

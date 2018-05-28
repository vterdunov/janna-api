package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/pkg/service"
)

// MakeInfoEndpoint returns an endpoint via the passed service
func MakeInfoEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		b, c := s.Info()
		return InfoResponse{b, c}, nil
	}
}

// InfoResponse is the Service build information
type InfoResponse struct {
	BuildTime string `json:"build_time"`
	Commit    string `json:"commit"`
}

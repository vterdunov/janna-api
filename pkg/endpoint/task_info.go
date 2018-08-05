package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/pkg/service"
)

// MakeTaskInfoEndpoint returns an endpoint via the passed service
func MakeTaskInfoEndpoint(s service.Service) endpoint.Endpoint { // nolint: dupl
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(TaskInfoRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		info, err := s.TaskInfo(ctx, req.TaskID)
		status := info.Status

		return TaskInfoResponse{Status: status, Err: err}, nil
	}
}

// TaskInfoRequest collects the request parameters for the TaskInfo method
type TaskInfoRequest struct {
	TaskID string
}

// TaskInfoResponse collects the response values for the TaskInfo method
type TaskInfoResponse struct {
	Err    error  `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
}

// Failed implements Failer
func (r TaskInfoResponse) Failed() error {
	return r.Err
}

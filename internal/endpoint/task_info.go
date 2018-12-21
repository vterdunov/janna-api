package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/vterdunov/janna-api/internal/service"
)

// MakeTaskInfoEndpoint returns an endpoint via the passed service
func MakeTaskInfoEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(TaskInfoRequest)
		if !ok {
			return nil, errors.New("could not parse request")
		}

		status, err := s.TaskInfo(ctx, req.TaskID)
		if err != nil {
			return TaskInfoResponse{Status: nil, Err: err}, nil
		}

		return TaskInfoResponse{Status: status, Err: err}, nil
	}
}

// TaskInfoRequest collects the request parameters for the TaskInfo method
type TaskInfoRequest struct {
	TaskID string
}

// TaskInfoResponse collects the response values for the TaskInfo method
type TaskInfoResponse struct {
	Status map[string]interface{}
	Err    error `json:"error,omitempty"`
}

// Failed implements Failer
func (r TaskInfoResponse) Failed() error {
	return r.Err
}

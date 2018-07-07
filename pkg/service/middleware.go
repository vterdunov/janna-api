package service

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"

	"github.com/vterdunov/janna-api/pkg/types"
)

// Middleware describes a service (as opposed to endpoint) middleware.
// type Middleware func(Service) Service

type loggingMiddleware struct {
	logger log.Logger
	Service
}

// // LoggingMiddleware takes a logger as a dependency
// // and returns a ServiceMiddleware.
// func LoggingMiddleware(logger log.Logger) Middleware {
// 	return func(next Service) Service {
// 		return loggingMiddleware{logger, next}
// 	}
// }

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingMiddleware{logger: logger, Service: s}
}

func (s *loggingMiddleware) VMList(ctx context.Context, params *types.VMListParams) (_ map[string]string, err error) {
	defer func() {
		s.logger.Log(
			"method", "VMList",
			"input", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMList(ctx, params)

}

// func (s *loggingMiddleware) FuncName() {}

func (s *loggingMiddleware) VMFind(ctx context.Context, params *types.VMFindParams) (_ *types.VMFound, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	s.logger.Log("request_id", reqID)

	defer func() {
		s.logger.Log(
			"method", "VMFind",
			"request_id", reqID,
			"input", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMFind(ctx, params)
}

func (s *loggingMiddleware) Info() (string, string) {
	defer func() {
		s.logger.Log(
			"method", "Info",
		)
	}()

	return s.Service.Info()
}

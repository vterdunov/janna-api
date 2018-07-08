package service

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"

	"github.com/vterdunov/janna-api/pkg/types"
)

type loggingMiddleware struct {
	logger log.Logger
	Service
}

// NewLoggingService returns a new instance of a logging Service.
// It used for business-domain logging.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingMiddleware{logger: logger, Service: s}
}

func (s *loggingMiddleware) VMList(ctx context.Context, params *types.VMListParams) (_ map[string]string, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)

	s.logger.Log(
		"msg", "calling method",
		"method", "VMList",
		"request_id", reqID,
		"params", fmt.Sprintf("%+v", params),
	)

	defer func() {
		s.logger.Log(
			"msg", "called method",
			"method", "VMList",
			"request_id", reqID,
			"err", err,
		)
	}()

	return s.Service.VMList(ctx, params)

}

func (s *loggingMiddleware) VMFind(ctx context.Context, params *types.VMFindParams) (_ *types.VMFound, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)

	defer func() {
		s.logger.Log(
			"method", "VMFind",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
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

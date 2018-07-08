package endpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// LoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
// This is a transport-domain logging
func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			logger.Log(
				"msg", "calling endpoint",
			)

			defer func(begin time.Time) {
				logger.Log(
					"msg", "called endpoint",
					"transport_error", err,
					"took", time.Since(begin))
			}(time.Now())

			return next(ctx, request)
		}
	}
}

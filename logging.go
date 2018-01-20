package main

import (
	"context"
	"time"

	log "github.com/rs/zerolog"
	"github.com/vterdunov/janna-api/types"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) VMInfo(ctx context.Context, name string) (_ types.VMSummary, err error) {
	defer func(begin time.Time) {
		mw.logger.Info().
			Str("method", "VMInfo").
			Dur("took", time.Since(begin)).
			Err(err)
	}(time.Now())
	return mw.next.VMInfo(ctx, name)
}

package main

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/vterdunov/janna-api/types"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// NewLoggingMiddleware create a new Logging Middleware
func NewLoggingMiddleware(logger log.Logger) Middleware {
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

func (mw loggingMiddleware) Info() (string, string) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "Info",
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Info()
}

func (mw loggingMiddleware) Healthz() bool {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "Healthz",
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Healthz()
}

func (mw loggingMiddleware) Readyz() bool {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "Readyz",
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Readyz()
}

func (mw loggingMiddleware) VMInfo(ctx context.Context, name string) (_ types.VMSummary, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "VMInfo",
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.VMInfo(ctx, name)
}

func (mw loggingMiddleware) VMDeploy(ctx context.Context, name string, OVAURL string, opts ...string) (_ int, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "VMInfo",
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.VMDeploy(ctx, name, OVAURL, opts...)
}

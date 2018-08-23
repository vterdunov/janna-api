package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"

	"github.com/vterdunov/janna-api/internal/types"
)

type instrumentingMiddleware struct {
	duration metrics.Histogram
	Service
}

// NewInstrumentingService returns a new instance of an instrumented Service.
// It used for business-domain instrumenting.
func NewInstrumentingService(duration metrics.Histogram) Middleware {
	return func(s Service) Service {
		return &instrumentingMiddleware{
			duration: duration,
			Service:  s,
		}
	}
}

func (mw instrumentingMiddleware) Info() (string, string) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Info", "success", "true"}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.Info()
}

func (mw *instrumentingMiddleware) VMList(ctx context.Context, params *types.VMListParams) (_ map[string]string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMList", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMList(ctx, params)
}

func (mw instrumentingMiddleware) VMInfo(ctx context.Context, params *types.VMInfoParams) (_ *types.VMSummary, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMInfo", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMInfo(ctx, params)
}

func (mw instrumentingMiddleware) VMFind(ctx context.Context, params *types.VMFindParams) (_ *types.VMFound, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMFind", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMFind(ctx, params)
}

func (mw instrumentingMiddleware) VMDeploy(ctx context.Context, params *types.VMDeployParams) (_ string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMDeploy", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMDeploy(ctx, params)
}

func (mw instrumentingMiddleware) VMSnapshotsList(ctx context.Context, params *types.VMSnapshotsListParams) (_ []types.Snapshot, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMSnapshotsList", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMSnapshotsList(ctx, params)
}

func (mw instrumentingMiddleware) VMSnapshotCreate(ctx context.Context, params *types.SnapshotCreateParams) (_ int32, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMSnapshotCreate", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMSnapshotCreate(ctx, params)
}

func (mw instrumentingMiddleware) VMRestoreFromSnapshot(ctx context.Context, params *types.VMRestoreFromSnapshotParams) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMRestoreFromSnapshot", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMRestoreFromSnapshot(ctx, params)
}

func (mw instrumentingMiddleware) VMRolesList(ctx context.Context, params *types.VMRolesListParams) (_ []types.Role, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMRolesList", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMRolesList(ctx, params)
}

func (mw instrumentingMiddleware) VMAddRole(ctx context.Context, params *types.VMAddRoleParams) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "VMAddRole", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.VMAddRole(ctx, params)
}

func (mw instrumentingMiddleware) RoleList(ctx context.Context) (_ []types.Role, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "RoleList", "success", fmt.Sprint(err == nil)}
		mw.duration.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Service.RoleList(ctx)
}

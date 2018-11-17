package service

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"

	"github.com/vterdunov/janna-api/internal/types"
)

type loggingMiddleware struct {
	logger log.Logger
	Service
}

// NewLoggingService returns a new instance of a logging Service.
// It used for business-domain logging.
func NewLoggingService(logger log.Logger) Middleware {
	return func(s Service) Service {
		return &loggingMiddleware{logger: logger, Service: s}
	}
}

func (s *loggingMiddleware) Info() (string, string) {
	defer func() {
		s.logger.Log(
			"method", "Info",
		)
	}()

	return s.Service.Info()
}

func (s *loggingMiddleware) VMList(ctx context.Context, params *types.VMListParams) (_ map[string]string, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMList",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMList(ctx, params)
}

func (s *loggingMiddleware) VMInfo(ctx context.Context, params *types.VMInfoParams) (_ *types.VMSummary, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMInfo",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMInfo(ctx, params)
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

func (s *loggingMiddleware) VMDelete(ctx context.Context, params *types.VMDeleteParams) (err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMDelete",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMDelete(ctx, params)
}

func (s *loggingMiddleware) VMDeploy(ctx context.Context, params *types.VMDeployParams) (_ string, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID).(string)
	defer func() {
		s.logger.Log(
			"method", "VMDeploy",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMDeploy(ctx, params)
}

func (s *loggingMiddleware) VMSnapshotsList(ctx context.Context, params *types.VMSnapshotsListParams) (_ []types.Snapshot, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMSnapshotsList",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMSnapshotsList(ctx, params)
}

func (s *loggingMiddleware) VMSnapshotCreate(ctx context.Context, params *types.SnapshotCreateParams) (_ int32, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMSnapshotCreate",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMSnapshotCreate(ctx, params)
}

func (s *loggingMiddleware) VMRestoreFromSnapshot(ctx context.Context, params *types.VMRestoreFromSnapshotParams) (err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMRestoreFromSnapshot",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMRestoreFromSnapshot(ctx, params)
}

func (s *loggingMiddleware) VMSnapshotDelete(ctx context.Context, params *types.VMSnapshotDeleteParams) (err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMSnapshotDelete",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMSnapshotDelete(ctx, params)
}

func (s *loggingMiddleware) VMPower(ctx context.Context, params *types.VMPowerParams) (err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"method", "VMPower",
			"request_id", reqID,
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMPower(ctx, params)
}

func (s *loggingMiddleware) VMRolesList(ctx context.Context, params *types.VMRolesListParams) (_ []types.Role, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"request_id", reqID,
			"method", "VMRolesList",
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMRolesList(ctx, params)
}

func (s *loggingMiddleware) VMAddRole(ctx context.Context, params *types.VMAddRoleParams) (err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"request_id", reqID,
			"method", "VMAddRole",
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMAddRole(ctx, params)
}

func (s *loggingMiddleware) VMScreenshot(ctx context.Context, params *types.VMScreenshotParams) (_ []byte, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"request_id", reqID,
			"method", "VMScreenshot",
			"params", fmt.Sprintf("%+v", params),
			"err", err,
		)
	}()

	return s.Service.VMScreenshot(ctx, params)
}

func (s *loggingMiddleware) RoleList(ctx context.Context) (_ []types.Role, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"request_id", reqID,
			"method", "RoleList",
			"err", err,
		)
	}()

	return s.Service.RoleList(ctx)
}

func (s *loggingMiddleware) TaskInfo(ctx context.Context, taskID string) (_ map[string]interface{}, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"request_id", reqID,
			"method", "TaskInfo",
			"err", err,
		)
	}()

	return s.Service.TaskInfo(ctx, taskID)
}

func (s *loggingMiddleware) OpenAPI(ctx context.Context) (_ []byte, err error) {
	reqID := ctx.Value(http.ContextKeyRequestXRequestID)
	defer func() {
		s.logger.Log(
			"request_id", reqID,
			"method", "OpenAPI",
			"err", err,
		)
	}()

	return s.Service.OpenAPI(ctx)
}

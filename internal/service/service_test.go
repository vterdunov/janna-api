package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/vmware/govmomi/vim25"
	"github.com/vterdunov/janna-api/internal/config"
	"github.com/vterdunov/janna-api/internal/types"
)

func TestNew(t *testing.T) {
	var svc Service
	type args struct {
		logger   log.Logger
		cfg      *config.Config
		client   *vim25.Client
		duration metrics.Histogram
		statuses Statuser
	}
	tests := []struct {
		name string
		args args
		want Service
	}{
		{name: "emptyService", args: args{}, want: svc},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.logger, tt.args.cfg, tt.args.client, tt.args.duration, tt.args.statuses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_newSimpleService(t *testing.T) {
	type args struct {
		logger   log.Logger
		cfg      *config.Config
		client   *vim25.Client
		statuses Statuser
	}
	tests := []struct {
		name string
		args args
		want Service
	}{
		{name: "emptyService", args: args{}, want: &service{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSimpleService(tt.args.logger, tt.args.cfg, tt.args.client, tt.args.statuses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSimpleService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_GetConfig(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	tests := []struct {
		name   string
		fields fields
		want   *config.Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			if got := s.GetConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_Info(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	tests := []struct {
		name   string
		fields fields
		want   string
		want1  string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, got1 := s.Info()
			if got != tt.want {
				t.Errorf("service.Info() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("service.Info() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_service_Healthz(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			if got := s.Healthz(); got != tt.want {
				t.Errorf("service.Healthz() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_Readyz(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			if got := s.Readyz(); got != tt.want {
				t.Errorf("service.Readyz() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMList(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMListParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.VMList(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.VMList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.VMList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMInfo(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMInfoParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.VMSummary
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.VMInfo(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.VMInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.VMInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMDelete(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMDeleteParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			if err := s.VMDelete(tt.args.ctx, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("service.VMDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_service_VMFind(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMFindParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.VMFound
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.VMFind(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.VMFind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.VMFind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMDeploy(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMDeployParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.VMDeploy(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.VMDeploy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("service.VMDeploy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMSnapshotsList(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMSnapshotsListParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.Snapshot
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.VMSnapshotsList(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.VMSnapshotsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.VMSnapshotsList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMSnapshotCreate(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.SnapshotCreateParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int32
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.VMSnapshotCreate(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.VMSnapshotCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("service.VMSnapshotCreate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMRestoreFromSnapshot(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMRestoreFromSnapshotParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			if err := s.VMRestoreFromSnapshot(tt.args.ctx, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("service.VMRestoreFromSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_service_VMSnapshotDelete(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMSnapshotDeleteParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			if err := s.VMSnapshotDelete(tt.args.ctx, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("service.VMSnapshotDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_service_VMRolesList(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMRolesListParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.Role
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.VMRolesList(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.VMRolesList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.VMRolesList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_VMAddRole(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		params *types.VMAddRoleParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			if err := s.VMAddRole(tt.args.ctx, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("service.VMAddRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_service_RoleList(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.Role
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.RoleList(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.RoleList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.RoleList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_TaskInfo(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		ctx    context.Context
		taskID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Task
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.TaskInfo(tt.args.ctx, tt.args.taskID)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.TaskInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.TaskInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_OpenAPI(t *testing.T) {
	type fields struct {
		logger   log.Logger
		cfg      *config.Config
		Client   *vim25.Client
		statuses Statuser
	}
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				logger:   tt.fields.logger,
				cfg:      tt.fields.cfg,
				Client:   tt.fields.Client,
				statuses: tt.fields.statuses,
			}
			got, err := s.OpenAPI(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.OpenAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.OpenAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

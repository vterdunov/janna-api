package status

import (
	"reflect"
	"testing"

	"github.com/vterdunov/janna-api/internal/service"
)

func TestNewStorage(t *testing.T) {
	s := NewStorage()
	tests := []struct {
		name string
		want *Storage
	}{
		{"simple storage", s},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_NewTask(t *testing.T) {
	st := NewStorage()
	task := st.NewTask()
	_ = task

	tests := []struct {
		name string
		s    *Storage
		want service.TaskStatuser
	}{
		// {"new task", st, task},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.NewTask(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.NewTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_FindByID(t *testing.T) {
	type args struct {
		id string
	}

	st := NewStorage()
	task := st.NewTask()
	id := task.ID()
	arg := args{id}

	tests := []struct {
		name string
		s    *Storage
		args args
		want service.TaskStatuser
	}{
		{"find", st, arg, task},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.FindByID(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.FindByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatus_ID(t *testing.T) {
	st := NewStorage()
	task := st.NewTask()
	ts := task.(*TaskStatus)
	id := ts.ID()

	tests := []struct {
		name string
		t    *TaskStatus
		want string
	}{
		{"id", ts, id},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.ID(); got != tt.want {
				t.Errorf("TaskStatus.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatus_Add(t *testing.T) {
	type args struct {
		keyvals []string
	}

	st := NewStorage()
	task := st.NewTask()
	ts := task.(*TaskStatus)

	arg := args{[]string{"key", "value"}}

	tests := []struct {
		name string
		t    *TaskStatus
		args args
	}{
		{"add", ts, arg},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.Str(tt.args.keyvals...)
		})
	}
}

func TestTaskStatus_Get(t *testing.T) {
	st := NewStorage()
	task := st.NewTask()
	fullTask := task.(*TaskStatus)
	fullTask.Str("key", "value")

	task2 := st.NewTask()
	mvTask := task2.(*TaskStatus)
	mvTask.Str("key")
	mMap := map[string]interface{}{"key": "(MISSING)"}

	tests := []struct {
		name         string
		t            *TaskStatus
		wantStatuses map[string]interface{}
	}{
		{"fullKeyValue", fullTask, fullTask.Get()},
		{"missingValue", mvTask, mMap},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotStatuses := tt.t.Get(); !reflect.DeepEqual(gotStatuses, tt.wantStatuses) {
				t.Errorf("TaskStatus.Get() = %v, want %v", gotStatuses, tt.wantStatuses)
			}
		})
	}
}

func TestStorage_gc(t *testing.T) {
	tests := []struct {
		name string
		s    *Storage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.gc()
		})
	}
}

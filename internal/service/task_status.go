package service

// Statuser represents behavior of storage that keeps statuses
// nolint: misspell
type Statuser interface {
	NewTask() TaskStatuser
	FindByID(id string) TaskStatuser
}

// TaskStatuser represents behavior of every single task
type TaskStatuser interface {
	ID() string
	Str(keyvals ...string) TaskStatuser
	StrArr(key string, arr []string) TaskStatuser
	Get() (statuses map[string]interface{})
}

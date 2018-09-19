package service

// Statuser represents how to get new storage to keep statuses
// nolint: misspell
type Statuser interface {
	NewTask() TaskStatuser
	FindByID(id string) TaskStatuser
}

// TaskStatuser represents behavior of every single task
type TaskStatuser interface {
	ID() string
	Add(statuses map[string]string)
	Get() (statuses map[string]string)
}

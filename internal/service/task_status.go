package service

// Statuser represents how to get new storage to keep statuses
type Statuser interface {
	NewTask() *TaskStatuser
}

// TaskStatuser represents behavoir of every single task
type TaskStatuser interface {
	Add(statuses map[string]string)
	Get() (statuses map[string]string)
}

package service

import "time"

type Task struct {
	Status     map[string]string
	Created    time.Time
	Expiration int64
}

type Statuser interface {
	Add(taskId string, payload map[string]string)
	Get(taskId string) *Task
}

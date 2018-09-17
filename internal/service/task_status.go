package service

import (
	"sync"
	"time"
)

type Task struct {
	sync.RWMutex
	ID string
	Status     map[string]string
	Created    time.Time
	Expiration int64
}

type Statuser interface {
	Add(taskId string, payload map[string]string)
	Get(taskId string) *Task
}

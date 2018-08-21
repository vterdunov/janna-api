package service

import "time"

type Task struct {
	Status     string
	Created    time.Time
	Expiration int64
}

type Statuser interface {
	Add(string, string)
	Get(string) *Task
}

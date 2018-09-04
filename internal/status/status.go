package status

import (
	"sync"
	"time"

	"github.com/vterdunov/janna-api/internal/service"
)

type Tasks struct {
	sync.RWMutex
	expiration    time.Duration
	cleanInterval time.Duration

	tasks map[string]service.Task
}

// New creates a new in-memory storage for tasks status information.
// Such as current progress, created date, etc.
func New() *Tasks {
	emptyTasks := make(map[string]service.Task)
	expiration := time.Hour * 24
	cleanInterval := time.Second * 10

	tt := Tasks{
		tasks:         emptyTasks,
		expiration:    expiration,
		cleanInterval: cleanInterval,
	}

	go tt.gc()

	return &tt
}

// Add a task to in-memory storage
func (tt *Tasks) Add(taskID string, Status map[string]string) {
	tt.Lock()
	defer tt.Unlock()

	expiration := time.Now().Add(tt.expiration).UnixNano()
	tt.tasks[taskID] = service.Task{
		Status:     Status,
		Created:    time.Now(),
		Expiration: expiration,
	}
}

// Get a task from in-memory storage
func (tt *Tasks) Get(taskID string) *service.Task {
	tt.RLock()
	defer tt.RUnlock()

	task, exists := tt.tasks[taskID]
	if !exists {
		return nil
	}

	isTaskExpired := time.Now().UnixNano() > task.Expiration
	if isTaskExpired {
		return nil
	}
	return &task
}

// gc search and clean expired tasks from in-memory storage
func (tt *Tasks) gc() {
	ticker := time.NewTicker(tt.cleanInterval)

	for range ticker.C {
		if tt.tasks == nil {
			return
		}
		var expiredTasksIDs []string

		tt.RLock()
		for id, task := range tt.tasks {
			isTaskExpired := time.Now().UnixNano() > task.Expiration
			if isTaskExpired {
				expiredTasksIDs = append(expiredTasksIDs, id)
			}
		}
		tt.RUnlock()

		for _, id := range expiredTasksIDs {
			delete(tt.tasks, id)
		}
	}
}

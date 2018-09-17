// status used to add and get information
// such as current deploy progress, deploy error messages, etc.
package status

import (
	"sync"
	"time"

	"github.com/vterdunov/janna-api/internal/service"
	"github.com/vterdunov/janna-api/pkg/uuid"
)

// Status stores information about something
type Status struct {
	sync.RWMutex
	expiration    time.Duration
	cleanInterval time.Duration

	records []*service.Task
}

// NewStorage creates a new in-memory storage
func NewStorage() *Status {
	expiration := time.Hour * 24
	cleanInterval := time.Second * 10

	tt := Status{
		expiration:    expiration,
		cleanInterval: cleanInterval,
	}

	go tt.gc()

	return &tt
}

// NewRecord creates a new unique status
func NewRecord() *service.Task {
	r := service.Task{
		ID: uuid.NewUUID(),
		Created:    time.Now(),
	}
	return &r
}

// Add a task to in-memory storage
func (t *service.Task) Add(msg map[string]string) {
	r.Lock()
	defer r.Unlock()

	expiration := time.Now().Add(r.expiration).UnixNano()
	tt.entry[taskID] = service.Task{
		Status:     Status,
		Created:    time.Now(),
		Expiration: expiration,
	}
}

// Get a task from in-memory storage
func (tt *Status) Get(taskID string) *service.Task {
	tt.RLock()
	defer tt.RUnlock()

	task, exists := tt.entry[taskID]
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
func (tt *Status) gc() {
	ticker := time.NewTicker(tt.cleanInterval)

	for range ticker.C {
		if tt.entry == nil {
			return
		}
		var expiredTasksIDs []string

		tt.RLock()
		for id, task := range tt.entry {
			isTaskExpired := time.Now().UnixNano() > task.Expiration
			if isTaskExpired {
				expiredTasksIDs = append(expiredTasksIDs, id)
			}
		}
		tt.RUnlock()

		for _, id := range expiredTasksIDs {
			delete(tt.entry, id)
		}
	}
}

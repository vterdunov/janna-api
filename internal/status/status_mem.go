// status used to add and get information
// such as current deploy progress, deploy error messages, etc.
package status

import (
	"fmt"
	"sync"
	"time"

	"github.com/vterdunov/janna-api/pkg/uuid"
)

// StatusStorage stores information about something
type StatusStorage struct {
	sync.RWMutex
	cleanInterval time.Duration

	tasks []TaskStatus
}

type TaskStatus struct {
	sync.RWMutex
	ID         string
	Status     map[string]string
	Created    time.Time
	expiration int64
}

// NewStorage creates a new in-memory storage
func NewStatusStorage() *StatusStorage {

	cleanInterval := time.Second * 10

	s := StatusStorage{
		cleanInterval: cleanInterval,
	}

	go s.gc()

	return &s
}

// NewTask creates a new unique status
func NewTask() *TaskStatus {
	expirationTime := time.Hour * 24

	expiration := time.Now().Add(expirationTime).UnixNano()
	r := TaskStatus{
		ID:         uuid.NewUUID(),
		Created:    time.Now(),
		expiration: expiration,
	}
	return &r
}

// Add a task to in-memory storage
func (t *TaskStatus) Add(statuses map[string]string) {
	t.Lock()
	defer t.Unlock()

	t.Status = statuses
}

// Get a task from in-memory storage
func (t *TaskStatus) Get() (statuses map[string]string) {
	t.RLock()
	defer t.RUnlock()

	// TODO: check status exist
	// task, exists := t.entry[taskID]
	// if !exists {
	// 	return nil
	// }

	return t.Status
}

// gc search and clean expired tasks from in-memory storage
func (s *StatusStorage) gc() {
	ticker := time.NewTicker(s.cleanInterval)

	for range ticker.C {
		if s.tasks == nil {
			return
		}
		// var expiredTasksIDs []string

		s.RLock()
		fmt.Println(s.tasks)
		for id, task := range s.tasks {
			isTaskExpired := time.Now().UnixNano() > task.expiration
			if isTaskExpired {
				s.tasks = append(s.tasks[:id], s.tasks[id+1:]...)
			}
		}
		s.RUnlock()
		fmt.Println(s.tasks)

		// for _, id := range expiredTasksIDs {
		// 	delete(s.tasks, id)
		// }
	}
}

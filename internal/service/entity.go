package service

import "time"

// VMSummary stores some information about Virtual Machines
type VMSummary struct {
	Name string
}

// VMUuid saves a VM uuid
type VMUuid struct {
	UUID string
	Name string
}

// Snapshot stores info about VM snapshot
type Snapshot struct {
	Name        string
	Description string
	ID          int32
	CreatedAt   time.Time
}

// Role represents ESXi role
type Role struct {
	Name        string
	ID          int32
	Description struct {
		Label   string
		Summary string
	}
}

package service

import "time"

// VMSummary stores some information about Virtual Machines
type VMSummary struct {
	Name             string
	Uuid             string
	Template         bool
	GuestId          string
	Annotation       string
	NumCpu           int32
	NumEthernetCards int32
	NumVirtualDisks  int32
}

// VMUuid saves a VM uuid
type VMUuid struct {
	Name string
	UUID string
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

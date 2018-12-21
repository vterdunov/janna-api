package domain

import "time"

// VMSummary stores some information about Virtual Machines
type VMSummary struct {
	BootTime            time.Time
	Name                string
	UUID                string
	GuestID             string
	Annotation          string
	PowerState          string
	ConnectionState     string
	NumCPU              int32
	NumEthernetCards    int32
	NumVirtualDisks     int32
	Paused              bool
	ConsolidationNeeded bool
	Template            bool
	VMGuestInfo
}

type VMGuestInfo struct {
	GuestID            string
	GuestFullName      string
	ToolsRunningStatus string
	HostName           string
	IPAddress          string
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

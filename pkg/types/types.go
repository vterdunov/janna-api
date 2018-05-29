package types

import (
	"time"

	vmwaretypes "github.com/vmware/govmomi/vim25/types"
)

// VMSummary stores some information about Virtual Machines
type VMSummary struct {
	Guest     *vmwaretypes.GuestInfo                  `json:"Guest,omitempty"`
	Heartbeat vmwaretypes.ManagedEntityStatus         `json:"HeartBeat,omitempty"`
	Runtime   vmwaretypes.VirtualMachineRuntimeInfo   `json:"Runtime,omitempty"`
	Config    vmwaretypes.VirtualMachineConfigSummary `json:"Config,omitempty"`
}

// Snapshot stores info about VM snapshot
type Snapshot struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ID          int32     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
}

// VMFound stores response from VMFind method
type VMFound struct {
	UUID string
	Name string
}

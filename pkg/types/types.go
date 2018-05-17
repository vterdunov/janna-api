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

// VMDeployParams stores VM deploy parameters like name, networks mapping and other
type VMDeployParams struct {
	Name       string
	OVAURL     string
	Datastores []string
	Networks   map[string]string
	Datacenter string
	Cluster    string
	Folder     string
}

type Snapshot struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ID          int32     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
}

type SnapshotCreateParams struct {
	VMName      string `json:"vm_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Memory      bool   `json:"memory"`
	Quiesce     bool   `json:"quiesce"`
}

type VMRestoreFromSnapshotParams struct {
	VMName  string `json:"vm_name"`
	Name    string `json:"name"`
	PowerOn bool   `json:"power_on"`
}

package types

import (
	"time"

	vmwaretypes "github.com/vmware/govmomi/vim25/types"

	"github.com/vterdunov/janna-api/pkg/config"
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

// Snapshot stores info about VM snapshot
type Snapshot struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ID          int32     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
}

// SnapshotCreateParams stores user request params
type SnapshotCreateParams struct {
	UUID        string `json:"vm_uuid"`
	Datacenter  string `json:"datacenter"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Memory      bool   `json:"memory"`
	Quiesce     bool   `json:"quiesce"`
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *SnapshotCreateParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

// VMRestoreFromSnapshotParams stores user request params
type VMRestoreFromSnapshotParams struct {
	UUID       string `json:"vm_uuid"`
	Datacenter string `json:"datacenter"`
	Name       string `json:"name"`
	PowerOn    bool   `json:"power_on"`
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMRestoreFromSnapshotParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

// VMInfoParams stores user request parameters
type VMInfoParams struct {
	UUID       string
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMInfoParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

// VMSnapshotsListParams stores user request parameters
type VMSnapshotsListParams struct {
	UUID       string
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMSnapshotsListParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

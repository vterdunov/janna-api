package types

import (
	"github.com/vterdunov/janna-api/pkg/config"
)

// VMListParams stores user request params
type VMListParams struct {
	Datacenter   string
	Folder       string
	ResourcePool string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMListParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}

	if p.Folder == "" {
		p.Folder = cfg.VMWare.Folder
	}
}

// VMDeployParams stores user request params
type VMDeployParams struct {
	Name       string
	OVAURL     string
	Datastores []string
	Networks   map[string]string
	Datacenter string
	Cluster    string
	Folder     string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMDeployParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}

	if p.Datastores == nil {
		// TODO: add default datastores
	}
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
	SnapshotID int32  `json:"snapshot_id"`
	Datacenter string `json:"datacenter"`
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

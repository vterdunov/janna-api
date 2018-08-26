package types

import (
	"github.com/vterdunov/janna-api/internal/config"
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

// SnapshotCreateParams stores user request params
type SnapshotCreateParams struct {
	UUID        string
	Datacenter  string
	Name        string
	Description string
	Memory      bool
	Quiesce     bool
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *SnapshotCreateParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

// VMRestoreFromSnapshotParams stores user request params
type VMRestoreFromSnapshotParams struct {
	UUID       string
	Datacenter string
	SnapshotID int32
	PowerOn    bool
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

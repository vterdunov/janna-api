package types

import "github.com/vterdunov/janna-api/internal/config"

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

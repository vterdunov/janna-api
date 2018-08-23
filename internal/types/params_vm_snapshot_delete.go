package types

import "github.com/vterdunov/janna-api/internal/config"

// VMSnapshotDeleteParams stores user request parameters
type VMSnapshotDeleteParams struct {
	UUID       string
	SnapshotID int32
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMSnapshotDeleteParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

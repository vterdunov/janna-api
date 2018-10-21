package types

import "github.com/vterdunov/janna-api/internal/config"

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

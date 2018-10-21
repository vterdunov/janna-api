package types

import "github.com/vterdunov/janna-api/internal/config"

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

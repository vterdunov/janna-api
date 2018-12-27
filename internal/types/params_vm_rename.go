package types

import "github.com/vterdunov/janna-api/internal/config"

// VMRenameParams stores user request parameters
type VMRenameParams struct {
	UUID       string
	Datacenter string
	Name       string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMRenameParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

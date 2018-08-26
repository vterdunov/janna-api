package types

import "github.com/vterdunov/janna-api/internal/config"

// VMDeleteParams stores user request parameters
type VMDeleteParams struct {
	UUID       string
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMDeleteParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

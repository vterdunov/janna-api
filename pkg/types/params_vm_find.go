package types

import "github.com/vterdunov/janna-api/pkg/config"

// VMFindParams stores user request parameters
type VMFindParams struct {
	Path       string
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMFindParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

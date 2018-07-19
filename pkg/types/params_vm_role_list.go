package types

import "github.com/vterdunov/janna-api/pkg/config"

// VMRolesListParams stores user request parameters
type VMRolesListParams struct {
	UUID       string
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMRolesListParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

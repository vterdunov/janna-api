package types

import "github.com/vterdunov/janna-api/internal/config"

// VMAddRoleParams stores user request parameters
type VMAddRoleParams struct {
	UUID       string
	Datacenter string
	Principal  string
	RoleID     int32
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMAddRoleParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

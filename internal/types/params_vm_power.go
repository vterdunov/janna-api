package types

import "github.com/vterdunov/janna-api/internal/config"

// VMPowerParams stores user request parameters
type VMPowerParams struct {
	UUID       string
	Datacenter string
	State      string
	Force      bool
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMPowerParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

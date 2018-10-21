package types

import (
	"github.com/vterdunov/janna-api/internal/config"
)

// VMInfoParams stores user request parameters
type VMInfoParams struct {
	UUID       string
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMInfoParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

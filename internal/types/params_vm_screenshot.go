package types

import "github.com/vterdunov/janna-api/internal/config"

// VMScreenshotParams stores user request parameters
type VMScreenshotParams struct {
	UUID       string
	Datacenter string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMScreenshotParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}
}

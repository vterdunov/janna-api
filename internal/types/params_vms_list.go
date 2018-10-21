package types

import "github.com/vterdunov/janna-api/internal/config"

// VMListParams stores user request params
type VMListParams struct {
	Datacenter   string
	Folder       string
	ResourcePool string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMListParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}

	if p.Folder == "" {
		p.Folder = cfg.VMWare.Folder
	}
}

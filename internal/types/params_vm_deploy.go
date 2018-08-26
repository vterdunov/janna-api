package types

import (
	"github.com/vterdunov/janna-api/internal/config"
)

// VMDeployParams stores user request params
type VMDeployParams struct {
	Name              string
	OVAURL            string
	Datacenter        string
	Folder            string
	Annotation        string
	Networks          map[string]string
	ComputerResources struct {
		Path string
		Type string
	}
	Datastores struct {
		Type  string
		Names []string
	}
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMDeployParams) FillEmptyFields(cfg *config.Config) {
	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}

	if p.Folder == "" {
		p.Folder = cfg.VMWare.Folder
	}

}

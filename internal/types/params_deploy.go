package types

import (
	"fmt"
	"strings"

	"github.com/vterdunov/janna-api/internal/config"
)

// VMDeployParams stores user request params
type VMDeployParams struct {
	Name              string
	OVAURL            string
	Datacenter        string
	Folder            string
	Datastores        []string
	Annotation        string
	Networks          map[string]string
	ComputerResources struct {
		Path string
		Type string
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

	if p.Datastores == nil {
		str := cfg.VMWare.DS
		ds := []string{}
		stores := strings.Split(str, ",")
		for _, store := range stores {
			ds = append(ds, strings.Trim(store, " "))
		}

		fmt.Println("--------------------")
		fmt.Println(ds)
		fmt.Println("--------------------")
		p.Datastores = ds
	}
}

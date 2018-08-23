package types

import (
	"fmt"
	"strings"

	"github.com/vterdunov/janna-api/internal/config"
)

// VMDeployParams stores user request params
type VMDeployParams struct {
	Name         string
	OVAURL       string
	Datastores   []string
	Networks     map[string]string
	Datacenter   string
	Host         string
	ResourcePool string
	Folder       string
	Annotation   string
}

// FillEmptyFields stores default parameters to the struct if some fields was empty
func (p *VMDeployParams) FillEmptyFields(cfg *config.Config) {
	fmt.Println("----->HERE")
	fmt.Println(p.Datastores)
	fmt.Println(cfg.VMWare.DS)

	if p.Datacenter == "" {
		p.Datacenter = cfg.VMWare.DC
	}

	if p.Folder == "" {
		p.Folder = cfg.VMWare.Folder
	}

	if p.ResourcePool == "" {
		p.ResourcePool = cfg.VMWare.RP
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

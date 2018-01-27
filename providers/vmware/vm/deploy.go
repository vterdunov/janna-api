package vm

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/vterdunov/janna-api/config"
)

// Deploy returns summary information about Virtual Machines
func Deploy(ctx context.Context, vmName string, OVAURL string, logger log.Logger, cfg *config.Config, opts ...string) (int, error) {
	var jid int
	// vmWareURL := cfg.Vmware.URL

	// u, err := soap.ParseURL(vmWareURL)
	// if err != nil {
	// 	logger.Log("err", "cannot parse VMWare URL")
	// 	return jid, err
	// }

	// insecure := cfg.Vmware.Insecure

	// c, err := govmomi.NewClient(ctx, u, insecure)
	// if err != nil {
	// 	logger.Log("err", err)
	// 	return jid, err
	// }

	// defer c.Logout(ctx)
	// f := find.NewFinder(c.Client, true)

	// dcName := cfg.Vmware.DC
	// dc, err := f.DatacenterOrDefault(ctx, dcName)
	// if err != nil {
	// 	logger.Log("err", err)
	// 	return jid, err
	// }

	// f.SetDatacenter(dc)
	logger.Log(
		"msg", "Deploy OVA",
		"name", vmName,
		"ova_url", OVAURL,
	)
	return jid, nil
}

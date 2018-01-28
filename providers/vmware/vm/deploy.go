package vm

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/vterdunov/janna-api/config"
)

// Deploy returns summary information about Virtual Machines
func Deploy(ctx context.Context, vmName string, OVAURL string, logger log.Logger, cfg *config.Config, opts ...string) (int, error) {
	// TODO: make up a metod to check deploy progress.
	// Job ID and endpoint with status?
	// keep HTTP connection with client and poll it?
	var jid int

	// 1) create empty struct that represents a deploy object

	// 2) Run a chain of calls:
	// 3) "Prepare":
	// 	 - validate OVA URL (try to use vim25.Client as Opener it has OpemRemote method)
	//   - fil OVA struct with: vim25.Client, Datacenter, Datastore, ResourcePool
	// 4) Download OVA OR use vim25.Client as Opener (see govc importx/archive.go:143)
	// 5) Import OVA (see govc importx)/ovf.go:212) it returns *types.ManagedObjectReference
	//   - Read file to []byte and unmarshal it to ovf.Envelope. The type allows to access to fileds of OVF strucure
	//   - Create a struct that represents a pert of OVF spec: name, nwtwork, etc.
	// 6) Create OVF Manager (ovf.NewManager)
	// 7) Fil/overrride the spec into current OVF (ovf.CreateImportSpec)
	// ?
	// 9) Get VM Folder.
	// 10) Crate Lease object (optional)
	// 11) Upload OVF
	// 12) ... Start

	logger.Log(
		"msg", "Deploy OVA",
		"name", vmName,
		"ova_url", OVAURL,
	)
	return jid, nil
}

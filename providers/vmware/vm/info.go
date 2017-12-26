package vmware

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// VMInfo returns slice of Virtual Machines
func VMInfo() {
	url := os.Getenv("VMWARE_URI")
	u, _ := soap.ParseURL(url)

	insecure := true
	ctx := context.Background()

	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		log.Print(err)
	}

	defer c.Logout(ctx)
	f := find.NewFinder(c.Client, true)

	dc, err := f.DatacenterOrDefault(ctx, "/DC1")

	if err != nil {
		log.Print(err)
	}

	f.SetDatacenter(dc)

	vms, err := f.VirtualMachineList(ctx, "coreos")
	if err != nil {
		log.Print(err)
	}

	pc := property.DefaultCollector(c.Client)

	// Retrieve summary property
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	props := []string{
		"summary",
		"guest.ipAddress",
		"config.extraConfig",
		"datastore",
		"network",
		"config.tools",
	}

	// Convert datastores into list of references
	var refs []types.ManagedObjectReference
	for _, vm := range vms {
		refs = append(refs, vm.Reference())
	}

	// Retrieve name property for all vms
	var vmt []mo.VirtualMachine
	err = pc.Retrieve(ctx, refs, props, &vmt)
	if err != nil {
		log.Fatal(err)
	}

	// Print name per virtual machine
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Println("Virtual machines found:", len(vmt))

	for _, vm := range vmt {
		fmt.Fprintf(tw, "%s\n", vm.Guest.IpAddress)
	}
	tw.Flush()
	// // Create view of VirtualMachine objects
	// m := view.NewManager(c.Client)

	// v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer v.Destroy(ctx)

	// // Retrieve summary property for all machines
	// // Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	// var vms []mo.VirtualMachine
	// refs := []string{"VirtualMachine"}
	// props := []string{
	// 	"summary",
	// 	"guest.ipAddress",
	// 	"config.extraConfig",
	// 	"datastore",
	// 	"network",
	// 	"config.tools",
	// }

	// err = v.Retrieve(ctx, refs, props, &vms)
	// if err != nil {
	// 	log.Fatal(err) // TODO: log the error and return 500
	// }

}

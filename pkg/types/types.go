package types

import vmwaretypes "github.com/vmware/govmomi/vim25/types"

// VMSummary stores some information about Virtual Machines
type VMSummary struct {
	// in: body
	Guest *vmwaretypes.GuestInfo `json:"Guest,omitempty"`
	// in: body
	Heartbeat vmwaretypes.ManagedEntityStatus `json:"HeartBeat,omitempty"`
	// in: body
	Runtime vmwaretypes.VirtualMachineRuntimeInfo `json:"Runtime,omitempty"`
	// in: body
	Config vmwaretypes.VirtualMachineConfigSummary `json:"Config,omitempty"`
}

// VMDeployParams stores VM deploy parameters like name, networks mapping and other
type VMDeployParams struct {
	Name       string
	OVAURL     string
	Datastores []string
	Networks   map[string]string
	Datacenter string
	Cluster    string
	Folder     string
}

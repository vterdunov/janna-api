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

// // AppBuildInfo stores the Service build information
// type AppBuildInfo struct {
// 	// in: body
// 	BuildTime string `json:"build_time"`
// 	// in: body
// 	Commit string `json:"commit"`
// }

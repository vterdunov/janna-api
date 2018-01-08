package vm

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/vmware/govmomi/vim25/types"
	vmware "github.com/vterdunov/janna-api/providers/vmware/vm"
)

type info struct {
	Guest     *types.GuestInfo                  `json:"Guest,omitempty"`
	Heartbeat types.ManagedEntityStatus         `json:"HeartBeat,omitempty"`
	Runtime   types.VirtualMachineRuntimeInfo   `json:"Runtime,omitempty"`
	Config    types.VirtualMachineConfigSummary `json:"Config,omitempty"`
}

// ReadInfo get information about VMs
func ReadInfo(w http.ResponseWriter, r *http.Request) {
	vmName := r.FormValue("vmname")
	inf, err := vmware.VMInfo(vmName)
	if err != nil {
		log.Error().Err(err).Msg("Couldn't get info about VM")
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	resp := &info{}
	for _, vmInfo := range inf {
		resp.Guest = vmInfo.Guest
		resp.Heartbeat = vmInfo.GuestHeartbeatStatus
		resp.Runtime = vmInfo.Summary.Runtime
		resp.Config = vmInfo.Summary.Config
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error().
			Err(err).
			Str("vmname", vmName).
			Msg("Couldn't encode info data")
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

package vm

import (
	"encoding/json"
	"log"
	"net/http"

	vmware "github.com/vterdunov/janna-api/providers/vmware/vm"
)

func ReadInfo(w http.ResponseWriter, r *http.Request) {
	vmName := r.FormValue("vmname")
	information, _ := vmware.VMInfo(vmName)

	body, err := json.Marshal(information)
	if err != nil {
		log.Printf("Couldn't encode info data: %v", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	// fmt.Fprint(w, information)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

package vm

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	vmware "github.com/vterdunov/janna-api/providers/vmware/vm"
)

func ReadInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	VMname := vars["name"]
	fmt.Println(VMname)

	vmware.VMInfo()

	// var vmStr string
	// for _, vm := range vms {
	// 	vmStr += fmt.Sprintf("%s: %s\n", vm.Summary.Config.Name, vm.Summary.Config.GuestFullName)
	// }
	// // vm.
	// fmt.Fprint(w, vmStr)
}

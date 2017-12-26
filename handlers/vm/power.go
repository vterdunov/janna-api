package vm

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func ReadPowerState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	VMname := vars["name"]
	fmt.Println(VMname)
}

func UpdatePowerState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	VMname := vars["name"]
	fmt.Println(VMname)
}

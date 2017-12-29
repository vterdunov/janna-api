package handlers

import (
	"github.com/gorilla/mux"
	"github.com/vterdunov/janna-api/handlers/health"
	"github.com/vterdunov/janna-api/handlers/vm"
)

// Router register necessary routes and returns an instance of a router.
func Router(buildTime, commit, release string) *mux.Router {
	r := mux.NewRouter()

	// Root
	r.HandleFunc("/", index).Methods("GET")

	// Version
	r.HandleFunc("/version", version(buildTime, commit, release)).Methods("GET")

	// Health
	r.HandleFunc("/healthz", health.Healthz).Methods("GET")
	r.HandleFunc("/readyz", health.Readyz).Methods("GET")

	// Test
	r.HandleFunc("/test/{name}/info", createTest).Methods("POST")
	r.HandleFunc("/test/{name}/info", readTest).Methods("GET")
	r.HandleFunc("/test/{name}/info", updateTest).Methods("UPDATE")
	r.HandleFunc("/test/{name}/info", deleteTest).Methods("DELETE")

	// VM
	r.HandleFunc("/vm/info", vm.ReadInfo).Methods("GET")

	r.HandleFunc("/vm/{name}/power", vm.ReadPowerState).Methods("GET")
	r.HandleFunc("/vm/{name}/power", vm.UpdatePowerState).Methods("UPDATE")

	return r
}

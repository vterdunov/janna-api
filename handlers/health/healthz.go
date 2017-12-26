package health

import "net/http"

// healthz is a liveness probe.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

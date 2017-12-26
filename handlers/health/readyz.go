package health

import (
	"net/http"
)

// readyz is a readiness probe.
func Readyz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

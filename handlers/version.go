package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// Application info
// swagger:response
type infoResponse struct {
	// in: body
	Payload *appInfo
}

type appInfo struct {
	BuildTime string `json:"build_time"`
	Commit    string `json:"commit"`
	Release   string `json:"release,omitempty"`
}

// getAppInfo swagger:route GET /info info
//
// return the Apllication version, commit and release info
//
// Responses:
//   200: infoResponse
func getAppInfo(buildTime, commit, release string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		info := appInfo{buildTime, commit, release}
		body, err := json.Marshal(info)
		if err != nil {
			log.Printf("Couldn't encode info data: %v", err)
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}
}

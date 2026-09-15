package httpserver

import "net/http"

type versionDTO struct {
	Version string `json:"version"`
}

func versionHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, versionDTO{Version: version})
	}
}

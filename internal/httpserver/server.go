// Package httpserver builds the pipeline-analytics HTTP handler tree.
package httpserver

import (
	"io/fs"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

// New returns the root HTTP handler. version is reported by GET
// /api/version -- pass the build's tagged version, or "dev" for a local
// build. assets is the built frontend (internal/webassets.FS()), served for
// every path the API doesn't claim, falling back to index.html for the SPA
// client router.
func New(registrar *ingestion.Registrar, store ingestion.Store, version string, assets fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/version", versionHandler(version))

	repos := &reposHandler{registrar: registrar, store: store}
	mux.HandleFunc("GET /api/repos", repos.list)
	mux.HandleFunc("POST /api/repos", repos.register)

	mux.Handle("/", staticHandler(assets))

	return mux
}

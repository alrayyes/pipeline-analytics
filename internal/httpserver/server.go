// Package httpserver builds the pipeline-analytics HTTP handler tree.
package httpserver

import (
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

// New returns the root HTTP handler.
func New(registrar *ingestion.Registrar, store ingestion.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	repos := &reposHandler{registrar: registrar, store: store}
	mux.HandleFunc("GET /api/repos", repos.list)
	mux.HandleFunc("POST /api/repos", repos.register)

	return mux
}

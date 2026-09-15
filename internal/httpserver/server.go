// Package httpserver builds the pipeline-analytics HTTP handler tree.
package httpserver

import "net/http"

// New returns the root HTTP handler.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return mux
}

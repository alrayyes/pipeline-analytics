package httpserver

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// staticHandler serves assets from an embedded SPA build, falling back to
// index.html for any path that isn't a real file -- the SvelteKit client
// router then handles it, the same shape as any other SPA fallback server.
func staticHandler(assets fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(assets))

	return func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" {
			clean = "."
		}

		info, err := fs.Stat(assets, clean)
		if err != nil || info.IsDir() {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	}
}

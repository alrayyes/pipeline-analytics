package httpserver

import (
	"encoding/json"
	"io/fs"
	"strings"
)

// routesFile is the dashboard's route list, written by web/'s build next to
// the pages (web/scripts/generate-routes.ts). It is how the server tells a
// client route from a mistyped URL.
const routesFile = "routes.json"

// routeSet matches a request path against the dashboard's route patterns,
// where a segment written [name] stands for any one non-empty segment.
type routeSet struct {
	patterns [][]string
}

// loadRoutes reads the route list from assets. It returns nil when there is
// none -- a checkout with no frontend build -- and the caller then serves the
// app shell for every path, as before the list existed.
func loadRoutes(assets fs.FS) *routeSet {
	if assets == nil {
		return nil
	}

	data, err := fs.ReadFile(assets, routesFile)
	if err != nil {
		return nil
	}

	var patterns []string
	if err := json.Unmarshal(data, &patterns); err != nil {
		return nil
	}

	routes := &routeSet{}
	for _, pattern := range patterns {
		routes.patterns = append(routes.patterns, segments(pattern))
	}

	return routes
}

// match reports whether name, a path with no leading slash, is a route.
func (r *routeSet) match(name string) bool {
	parts := segments(name)

next:
	for _, pattern := range r.patterns {
		if len(pattern) != len(parts) {
			continue
		}

		for i, want := range pattern {
			dynamic := strings.HasPrefix(want, "[") && strings.HasSuffix(want, "]")
			if parts[i] == "" || (!dynamic && parts[i] != want) {
				continue next
			}
		}

		return true
	}

	return false
}

// segments splits a path into its non-root segments: "/" and "" are none.
func segments(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}

	return strings.Split(p, "/")
}

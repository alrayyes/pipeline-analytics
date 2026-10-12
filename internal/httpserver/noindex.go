package httpserver

import "net/http"

// noindex tells search engines to keep every response out of their index and
// not to follow its links. This is an internal tool (alrayyes/dotfiles#810).
// It is wrapped outermost, so a request rejected before any handler runs, the
// health probes and the static files carry it too. robots.txt must stay
// permissive: a Disallow would stop a crawler fetching the response, and it
// would never read this.
func noindex(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}

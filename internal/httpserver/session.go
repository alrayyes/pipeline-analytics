package httpserver

import (
	"net/http"
	"strings"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
)

// requireSession gates every request behind a valid session, except the
// WebAuthn ceremony endpoints, the forge webhook receivers, and the static
// frontend -- per dashboard-auth/spec.md's "Session-gated access". The
// static SPA shell has to stay reachable unauthenticated so its own login
// page can load at all; the data APIs behind it are what's actually gated.
func requireSession(store auth.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)

			return
		}

		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "no valid session")

			return
		}

		if _, err := store.Session(r.Context(), cookie.Value); err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "no valid session")

			return
		}

		next.ServeHTTP(w, r)
	})
}

func isPublicPath(path string) bool {
	switch {
	case path == "/healthz", path == "/api/version":
		return true
	case strings.HasPrefix(path, "/api/auth/"):
		return true
	case strings.HasPrefix(path, "/webhooks/"):
		return true
	case !strings.HasPrefix(path, "/api/"):
		return true // the static frontend shell
	default:
		return false
	}
}

package httpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
)

// authInfo is stashed on the request context by requireSession, so a
// handler downstream can tell who's authenticated and how -- see
// add-api-token-auth/design.md's "session-only" decision for why the
// "how" matters to the token endpoints specifically.
type authInfo struct {
	UserID     string
	ViaSession bool
}

type authInfoKey struct{}

func withAuthInfo(r *http.Request, info authInfo) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authInfoKey{}, info))
}

// authInfoFromContext returns the authInfo requireSession stashed on r, if
// any. Only false on a path requireSession let through unauthenticated
// (isPublicPath) -- every gated path always has one by the time a handler
// runs.
func authInfoFromContext(r *http.Request) (authInfo, bool) {
	info, ok := r.Context().Value(authInfoKey{}).(authInfo)

	return info, ok
}

// requireSession gates every request behind a valid session or API token,
// except the WebAuthn ceremony endpoints, the forge webhook receivers, the
// static frontend, and -- specifically requiring a session, not a token --
// token issuance/revocation, per dashboard-auth/spec.md's "Session-gated
// access". The static SPA shell has to stay reachable unauthenticated so
// its own login page can load at all; the data APIs behind it are what's
// actually gated.
func requireSession(store auth.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)

			return
		}

		if cookie, err := r.Cookie(sessionCookieName); err == nil {
			if userID, err := store.Session(r.Context(), cookie.Value); err == nil {
				next.ServeHTTP(w, withAuthInfo(r, authInfo{UserID: userID, ViaSession: true}))

				return
			}
		}

		if token, ok := bearerToken(r); ok {
			if userID, err := store.TokenUserID(r.Context(), token); err == nil {
				next.ServeHTTP(w, withAuthInfo(r, authInfo{UserID: userID, ViaSession: false}))

				return
			}
		}

		writeError(w, http.StatusUnauthorized, "unauthenticated", "no valid session or token")
	})
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "

	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return "", false
	}

	token := strings.TrimPrefix(auth, prefix)
	if token == "" {
		return "", false
	}

	return token, true
}

func isPublicPath(path string) bool {
	switch {
	case path == "/healthz", path == "/api/version":
		return true
	case path == "/api/auth/tokens", strings.HasPrefix(path, "/api/auth/tokens/"):
		return false // session-only, not the general /api/auth/ exemption below
	case path == "/api/auth/credentials", strings.HasPrefix(path, "/api/auth/credentials/"):
		return false // session-only, same reasoning as tokens above
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

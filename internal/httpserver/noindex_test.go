package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// An internal tool must not be indexed (alrayyes/dotfiles#810), so every
// response says so: the app shell, JSON, health probes, static files, errors.
func TestEveryResponseAsksNotToBeIndexed(t *testing.T) {
	t.Parallel()

	paths := []string{
		"/",
		"/healthz",
		"/readyz",
		"/api/version",
		"/api/repos",
		"/api/no-such-endpoint",
		"/robots.txt",
		"/no/such/page",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			newTestServer(t, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

			require.Equal(t, "noindex, nofollow", rec.Header().Get("X-Robots-Tag"))
		})
	}
}

func TestRejectedRequestsAskNotToBeIndexedToo(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	newTestServer(t, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/repos", nil))

	require.GreaterOrEqual(t, rec.Code, http.StatusBadRequest)
	require.Equal(t, "noindex, nofollow", rec.Header().Get("X-Robots-Tag"))
}

func TestRobotsTxtDoesNotDisallowCrawling(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	newTestServer(t, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	// A Disallow: / would keep a crawler from fetching the page, so it would
	// never read the noindex.
	require.NotContains(t, rec.Body.String(), "Disallow: /\n")
}

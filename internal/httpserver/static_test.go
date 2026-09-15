package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestStaticServing(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{
		"index.html":   {Data: []byte("<html>index</html>")},
		"_app/main.js": {Data: []byte("console.log('hi')")},
	}

	t.Run("serves a real file directly", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/_app/main.js", nil)
		rec := httptest.NewRecorder()
		newTestServerWithAssets(t, nil, assets).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "console.log('hi')", rec.Body.String())
	})

	t.Run("falls back to index.html for an unmatched client route", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/releases", nil)
		rec := httptest.NewRecorder()
		newTestServerWithAssets(t, nil, assets).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<html>index</html>", rec.Body.String())
	})

	t.Run("serves index.html at the root", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		newTestServerWithAssets(t, nil, assets).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<html>index</html>", rec.Body.String())
	})

	t.Run("api routes still take priority over the static catch-all", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		rec := httptest.NewRecorder()
		newTestServerWithAssets(t, nil, assets).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "test-version")
	})
}

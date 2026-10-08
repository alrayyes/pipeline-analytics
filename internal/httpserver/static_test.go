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

func TestStaticCaching(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{
		"index.html":                  {Data: []byte("<html>index</html>")},
		"manifest.json":               {Data: []byte("{}")},
		"_app/immutable/chunk.abc.js": {Data: []byte("console.log('hi')")},
	}

	cacheControl := func(t *testing.T, target string) string {
		t.Helper()

		req := httptest.NewRequest(http.MethodGet, target, nil)
		rec := httptest.NewRecorder()
		newTestServerWithAssets(t, nil, assets).ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		return rec.Header().Get("Cache-Control")
	}

	t.Run("a hashed asset is cached for a year and never revalidated", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "public, max-age=31536000, immutable", cacheControl(t, "/_app/immutable/chunk.abc.js"))
	})

	t.Run("the page is revalidated", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "no-cache", cacheControl(t, "/"))
	})

	t.Run("a client route served as the page is revalidated", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "no-cache", cacheControl(t, "/releases"))
	})

	t.Run("an unhashed static file is revalidated", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "no-cache", cacheControl(t, "/manifest.json"))
	})
}

func TestStaticCompression(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{
		"index.html":                   {Data: []byte("<html>index</html>")},
		"index.html.br":                {Data: []byte("br-index")},
		"_app/immutable/app.abc.js":    {Data: []byte("plain-js")},
		"_app/immutable/app.abc.js.br": {Data: []byte("br-js")},
		"_app/immutable/app.abc.js.gz": {Data: []byte("gz-js")},
		"icon.png":                     {Data: []byte("png")},
	}

	get := func(t *testing.T, target, acceptEncoding string) *httptest.ResponseRecorder {
		t.Helper()

		req := httptest.NewRequest(http.MethodGet, target, nil)
		if acceptEncoding != "" {
			req.Header.Set("Accept-Encoding", acceptEncoding)
		}
		rec := httptest.NewRecorder()
		newTestServerWithAssets(t, nil, assets).ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		return rec
	}

	t.Run("serves brotli to a client that takes it", func(t *testing.T) {
		t.Parallel()

		rec := get(t, "/_app/immutable/app.abc.js", "gzip, br")

		require.Equal(t, "br-js", rec.Body.String())
		require.Equal(t, "br", rec.Header().Get("Content-Encoding"))
	})

	t.Run("serves gzip when brotli is not accepted", func(t *testing.T) {
		t.Parallel()

		rec := get(t, "/_app/immutable/app.abc.js", "gzip")

		require.Equal(t, "gz-js", rec.Body.String())
		require.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	})

	t.Run("serves the plain file to a client that takes neither", func(t *testing.T) {
		t.Parallel()

		rec := get(t, "/_app/immutable/app.abc.js", "")

		require.Equal(t, "plain-js", rec.Body.String())
		require.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("keeps the original file's content type", func(t *testing.T) {
		t.Parallel()

		rec := get(t, "/_app/immutable/app.abc.js", "br")

		require.Contains(t, rec.Header().Get("Content-Type"), "javascript")
	})

	t.Run("tells caches the response depends on the encoding", func(t *testing.T) {
		t.Parallel()

		rec := get(t, "/_app/immutable/app.abc.js", "br")

		require.Equal(t, "Accept-Encoding", rec.Header().Get("Vary"))
	})

	t.Run("serves the page compressed too", func(t *testing.T) {
		t.Parallel()

		rec := get(t, "/", "br")

		require.Equal(t, "br-index", rec.Body.String())
	})

	t.Run("serves a file with no compressed copy as it is", func(t *testing.T) {
		t.Parallel()

		rec := get(t, "/icon.png", "br, gzip")

		require.Equal(t, "png", rec.Body.String())
	})
}

func TestStaticRevalidation(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{"index.html": {Data: []byte("<html>index</html>")}}
	srv := newTestServerWithAssets(t, nil, assets)

	first := httptest.NewRecorder()
	srv.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	etag := first.Header().Get("ETag")
	require.NotEmpty(t, etag)

	t.Run("answers a matching ETag with 304 and no body", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("If-None-Match", etag)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotModified, rec.Code)
	})
}

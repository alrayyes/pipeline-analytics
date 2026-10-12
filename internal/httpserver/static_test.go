package httpserver_test

import (
	"errors"
	"io/fs"
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

const routesJSON = `["/", "/runs", "/runs/[id]", "/login"]`

func errorAssets() fstest.MapFS {
	return fstest.MapFS{
		"index.html":                  {Data: []byte("<html>shell</html>")},
		"404.html":                    {Data: []byte("<html>Page not found</html>")},
		"500.html":                    {Data: []byte("<html>Something went wrong</html>")},
		"routes.json":                 {Data: []byte(routesJSON)},
		"_app/immutable/chunk.abc.js": {Data: []byte("console.log('hi')")},
	}
}

func TestStaticErrorStatus(t *testing.T) {
	t.Parallel()

	get := func(t *testing.T, assets fs.FS, target string) *httptest.ResponseRecorder {
		t.Helper()

		req := httptest.NewRequest(http.MethodGet, target, nil)
		rec := httptest.NewRecorder()
		newTestServerWithAssets(t, nil, assets).ServeHTTP(rec, req)

		return rec
	}

	t.Run("a path no route matches answers 404 with the app shell, which renders the not-found view", func(t *testing.T) {
		t.Parallel()

		rec := get(t, errorAssets(), "/no/such/page")

		require.Equal(t, http.StatusNotFound, rec.Code)
		require.Equal(t, "<html>shell</html>", rec.Body.String())
		require.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		require.Equal(t, "noindex, nofollow", rec.Header().Get("X-Robots-Tag"))
		require.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	})

	t.Run("a route still gets the app shell", func(t *testing.T) {
		t.Parallel()

		for _, target := range []string{"/", "/runs", "/runs/", "/login"} {
			rec := get(t, errorAssets(), target)

			require.Equal(t, http.StatusOK, rec.Code, target)
			require.Equal(t, "<html>shell</html>", rec.Body.String(), target)
		}
	})

	t.Run("a dynamic route matches one segment", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, http.StatusOK, get(t, errorAssets(), "/runs/42").Code)
		require.Equal(t, http.StatusNotFound, get(t, errorAssets(), "/runs/42/extra").Code)
	})

	t.Run("a file that doesn't exist answers 404, not the shell", func(t *testing.T) {
		t.Parallel()

		rec := get(t, errorAssets(), "/_app/immutable/gone.js")

		require.Equal(t, http.StatusNotFound, rec.Code)
		require.NotContains(t, rec.Body.String(), "shell")
	})

	t.Run("an unknown API path keeps the JSON error shape", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithAssets(t, nil, errorAssets())
		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/no-such-endpoint", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
		require.JSONEq(t, `{"code":"not_found","message":"no such endpoint"}`, rec.Body.String())
	})

	t.Run("a 404 with no app shell falls back to the static not-found page", func(t *testing.T) {
		t.Parallel()

		assets := errorAssets()
		delete(assets, "index.html")

		rec := get(t, assets, "/no/such/page")

		require.Equal(t, http.StatusNotFound, rec.Code)
		require.Equal(t, "<html>Page not found</html>", rec.Body.String())
	})

	t.Run("a 404 with neither page still answers 404", func(t *testing.T) {
		t.Parallel()

		assets := errorAssets()
		delete(assets, "index.html")
		delete(assets, "404.html")

		rec := get(t, assets, "/no/such/page")

		require.Equal(t, http.StatusNotFound, rec.Code)
		require.NotEmpty(t, rec.Body.String())
	})

	t.Run("a build with no route list keeps serving the shell", func(t *testing.T) {
		t.Parallel()

		assets := errorAssets()
		delete(assets, "routes.json")

		rec := get(t, assets, "/no/such/page")

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<html>shell</html>", rec.Body.String())
	})

	t.Run("an unreadable asset answers 500 with the error page and no detail", func(t *testing.T) {
		t.Parallel()

		assets := brokenFS{FS: errorAssets(), broken: "index.html"}

		rec := get(t, assets, "/runs")

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Equal(t, "<html>Something went wrong</html>", rec.Body.String())
		require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
		require.NotContains(t, rec.Body.String(), "disk on fire")
	})
}

// brokenFS lets a file be found but fails reading it, the way a damaged
// embedded asset would.
type brokenFS struct {
	fs.FS

	broken string
}

func (b brokenFS) Open(name string) (fs.File, error) {
	file, err := b.FS.Open(name)
	if err != nil || name != b.broken {
		return file, err //nolint:wrapcheck // a pass-through of the wrapped FS
	}

	return brokenFile{File: file}, nil
}

var errDiskOnFire = errors.New("disk on fire")

type brokenFile struct{ fs.File }

func (brokenFile) Read([]byte) (int, error) { return 0, errDiskOnFire }

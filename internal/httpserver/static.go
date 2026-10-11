package httpserver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	// SvelteKit writes content-hashed files under _app/immutable/, so their
	// name changes when their bytes do and they can be kept for good.
	immutablePrefix = "_app/immutable/"
	immutableCache  = "public, max-age=31536000, immutable"
	// Everything else keeps its URL across builds, so a cache has to ask
	// whether it is still current (the ETag answers with a 304).
	revalidateCache = "no-cache"
)

// encodings are the precompressed copies adapter-static writes beside each
// text file, in the order they are preferred.
var encodings = []string{"br", "gzip"}

var encodingSuffix = map[string]string{"br": ".br", "gzip": ".gz"}

type staticFile struct {
	data []byte
	etag string
}

// staticHandler serves assets from an embedded SPA build. A path that isn't a
// file but is one of the dashboard's routes (routes.json) gets index.html, and
// the SvelteKit client router takes over. Anything else answers 404, so a
// mistyped URL isn't a soft 404 (rules/error-pages.md), but still with the app
// shell: the router finds no route and renders the not-found view inside the
// dashboard's own layout, which a static page couldn't. A file the server can't
// read answers 500.
//
// Hashed assets are cached for a year and the rest revalidated, and a file
// with a .br or .gz copy next to it is served that way to a client that
// accepts it, so nothing is compressed at request time.
func staticHandler(assets fs.FS) http.HandlerFunc {
	load := assetLoader(assets)
	routes := loadRoutes(assets)

	return func(w http.ResponseWriter, r *http.Request) {
		// Only reached when no API route matched, so this is a typo, and an
		// API client expects the API's error shape, not a page.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "not_found", "no such endpoint")

			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if info, err := fs.Stat(assets, name); err != nil || info.IsDir() {
			// The app's own files never fall back: a deleted hashed asset is
			// a plain 404, not an HTML page the browser then fails to run.
			if strings.HasPrefix(name, "_app/") {
				w.Header().Set("Cache-Control", revalidateCache)
				http.NotFound(w, r)

				return
			}

			if routes != nil && !routes.match(name) {
				serveError(w, r, assets, load, http.StatusNotFound)

				return
			}

			name = "index.html"
		}

		served, encoding := negotiate(assets, name, r.Header.Get("Accept-Encoding"))

		file, err := load(served)
		if err != nil {
			slog.ErrorContext(r.Context(), "read embedded asset", "asset", served, "error", err)
			serveError(w, r, assets, load, http.StatusInternalServerError)

			return
		}

		setAssetHeaders(w.Header(), name, encoding, file.etag)
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(file.data))
	}
}

// errorPage is how the handler answers one error status: the files to try in
// order (the first it can read wins), the plain-text body for a build with
// none of them, and the cache policy. The static pages are the ones the web
// build writes, self-contained so they still render when the app can't.
type errorPage struct {
	files []string
	text  string
	cache string
}

var errorPages = map[int]errorPage{
	// The shell first, so the not-found view renders in the dashboard layout.
	http.StatusNotFound:            {[]string{"index.html", "404.html"}, "Page not found", revalidateCache},
	http.StatusInternalServerError: {[]string{"500.html"}, "Something went wrong", "no-store"},
}

// serveError answers with status and its page. Nothing about the failure is
// sent: the caller logged it.
func serveError(w http.ResponseWriter, r *http.Request, assets fs.FS, load func(string) (staticFile, error), status int) {
	page := errorPages[status]

	for _, name := range page.files {
		served, encoding := negotiate(assets, name, r.Header.Get("Accept-Encoding"))

		file, err := load(served)
		if err != nil {
			continue
		}

		setAssetHeaders(w.Header(), name, encoding, file.etag)
		w.Header().Del("ETag")
		w.Header().Set("Cache-Control", page.cache)
		w.WriteHeader(status)

		if r.Method != http.MethodHead {
			_, _ = w.Write(file.data)
		}

		return
	}

	w.Header().Set("Cache-Control", page.cache)
	http.Error(w, page.text, status)
}

// assetLoader reads an asset once and keeps it with its ETag, since the
// embedded files can't change while the process runs.
func assetLoader(assets fs.FS) func(name string) (staticFile, error) {
	var cache sync.Map // asset name -> staticFile

	return func(name string) (staticFile, error) {
		if cached, ok := cache.Load(name); ok {
			file, _ := cached.(staticFile)

			return file, nil
		}

		data, err := fs.ReadFile(assets, name)
		if err != nil {
			return staticFile{}, fmt.Errorf("read %s: %w", name, err)
		}

		sum := sha256.Sum256(data)
		file := staticFile{data: data, etag: `"` + hex.EncodeToString(sum[:8]) + `"`}
		cache.Store(name, file)

		return file, nil
	}
}

// negotiate picks the precompressed copy of name that the client accepts, or
// name itself with no encoding when there is none.
func negotiate(assets fs.FS, name, acceptEncoding string) (served, encoding string) {
	for _, candidate := range encodings {
		if !acceptsEncoding(acceptEncoding, candidate) {
			continue
		}
		if _, err := fs.Stat(assets, name+encodingSuffix[candidate]); err == nil {
			return name + encodingSuffix[candidate], candidate
		}
	}

	return name, ""
}

func setAssetHeaders(header http.Header, name, encoding, etag string) {
	header.Add("Vary", "Accept-Encoding")
	header.Set("ETag", etag)
	header.Set("Cache-Control", revalidateCache)
	if strings.HasPrefix(name, immutablePrefix) {
		header.Set("Cache-Control", immutableCache)
	}
	if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
		header.Set("Content-Type", contentType)
	}
	if encoding != "" {
		header.Set("Content-Encoding", encoding)
	}
}

// acceptsEncoding reports whether an Accept-Encoding header lists the coding.
// Quality values are ignored beyond an explicit q=0: every client that sends
// br or gzip wants them, and none ranks them against each other here.
func acceptsEncoding(header, coding string) bool {
	for part := range strings.SplitSeq(header, ",") {
		token, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if !strings.EqualFold(strings.TrimSpace(token), coding) {
			continue
		}

		return strings.ReplaceAll(strings.TrimSpace(params), " ", "") != "q=0"
	}

	return false
}

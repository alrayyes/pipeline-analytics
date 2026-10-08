package httpserver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
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

// staticHandler serves assets from an embedded SPA build, falling back to
// index.html for any path that isn't a real file -- the SvelteKit client
// router then handles it, the same shape as any other SPA fallback server.
//
// Hashed assets are cached for a year and the rest revalidated, and a file
// with a .br or .gz copy next to it is served that way to a client that
// accepts it, so nothing is compressed at request time.
func staticHandler(assets fs.FS) http.HandlerFunc {
	load := assetLoader(assets)

	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if info, err := fs.Stat(assets, name); err != nil || info.IsDir() {
			name = "index.html"
		}

		served, encoding := negotiate(assets, name, r.Header.Get("Accept-Encoding"))

		file, ok := load(served)
		if !ok {
			http.NotFound(w, r)

			return
		}

		setAssetHeaders(w.Header(), name, encoding, file.etag)
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(file.data))
	}
}

// assetLoader reads an asset once and keeps it with its ETag, since the
// embedded files can't change while the process runs.
func assetLoader(assets fs.FS) func(name string) (staticFile, bool) {
	var cache sync.Map // asset name -> staticFile

	return func(name string) (staticFile, bool) {
		if cached, ok := cache.Load(name); ok {
			file, _ := cached.(staticFile)

			return file, true
		}

		data, err := fs.ReadFile(assets, name)
		if err != nil {
			return staticFile{}, false
		}

		sum := sha256.Sum256(data)
		file := staticFile{data: data, etag: `"` + hex.EncodeToString(sum[:8]) + `"`}
		cache.Store(name, file)

		return file, true
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

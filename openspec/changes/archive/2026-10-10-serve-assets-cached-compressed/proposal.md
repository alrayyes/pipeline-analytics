# Proposal

## Why

Lighthouse reports the dashboard's CSS and JS with no cache lifetime and no
compression, because the Go server sends them with neither. Visitors
re-download unchanged assets, and the first load carries 80 KB of uncompressed
CSS. Tracked in #523, applying the web-performance rule.

## What Changes

- Files under `_app/immutable/` are sent with
  `Cache-Control: public, max-age=31536000, immutable`. Everything else, the
  page included, is sent with `no-cache` and an `ETag`, and a matching
  `If-None-Match` gets a 304.
- The build writes `.br` and `.gz` copies of its text files
  (`adapter-static`'s `precompress`), and the server serves one by
  `Accept-Encoding`, with `Vary: Accept-Encoding` and the original file's
  content type.
- `lighthouserc.cjs` asserts the four insight IDs at `warn`:
  `cache-insight`, `document-latency-insight`, `render-blocking-insight` and
  `network-dependency-tree-insight`. The first two now pass; the last two don't
  (see Out of scope).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `deployment`: adds a requirement for how the server delivers its assets.

## Impact

- `internal/httpserver/static.go` and its tests, `web/vite.config.ts`,
  `web/lighthouserc.cjs`, `README.md`.
- `dist/` grows by the compressed copies, which are build output and ignored.
- The header and encoding tests in `static_test.go` are the header test the
  rule asks for; Lighthouse already runs against this server.

## Out of scope

- Render-blocking CSS: one 80 KB stylesheet, all of it used across the app.
  Inlining it would move the bytes into every page; splitting it per route is
  its own ticket.
- The stylesheet-to-font request chain. Font file names are hashed per build,
  so a preload needs build support.
- Compressing API responses, which belongs at the proxy.

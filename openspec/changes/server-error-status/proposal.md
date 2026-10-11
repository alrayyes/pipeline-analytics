## Why

The server answers every path that isn't a file with `200` and the app shell,
so a mistyped URL is a soft 404: Google flags a 2xx page that reads like an
error, and an uptime check can't tell a missing page from a working one.
`rules/error-pages.md` wants a real `404` and `500` with the page in the site's
own layout. The pages are #576; the status codes and cache headers are #577.

## What Changes

- The server knows the dashboard's routes from a `routes.json` the web build
  writes next to the pages, and answers `404` with `404.html` for a path that
  matches none of them and isn't a file.
- A request for a file that doesn't exist (a deleted hashed asset, for example)
  answers `404`, not the app shell.
- A failure to read an embedded asset answers `500` with `500.html` and logs the
  detail without sending it.
- Error pages are sent with `Cache-Control: no-cache` (404) or `no-store` (500).
- `/api/` paths keep their JSON error shape.
- A build with no `routes.json` (a bare `go test` checkout) keeps today's
  behaviour, so the Go tests don't need a frontend build.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `deployment`: adds a requirement for the status and caching of error
  responses next to the existing asset-caching one.

## Impact

- `internal/httpserver/static.go` and its tests, a new routes matcher beside it.
- `web/scripts/generate-routes.ts` and `web/package.json` (`postbuild`).
- `README.md`.

# Proposal

## Why

The dashboard had no error views: SvelteKit's default output showed for an
unknown path or a failed load. `rules/error-pages.md` asks for a 404 always and
a 500 wherever something runs on the server, in the site's own layout, in plain
words, with a way out and no internals. Tracked in #576. The server half (real
status codes, `Cache-Control`) is #577.

## What Changes

- `+error.svelte`: "Page not found" for a 404, "Something went wrong" for
  anything else, in the shared layout, with a status chip in words, one
  sentence and a "Back to the overview" action. The wording depends on the
  status only; the error's message is never rendered.
- `handleError` on the server and client hooks logs the detail and returns a
  generic message.
- `src/error.html` for errors in `handle` and the root layout, and
  `static/404.html` and `static/500.html` as self-contained pages the build
  puts in the embedded `dist/`, for the server to fall back on.
- Adopted from the Stitch candidate (screen "Pipeline Error Diagnostics"): the
  chip, the large headline over a short explanation, the full-width primary
  action. Rejected: the stderr block, trace ID, cluster, service and version
  details, error codes as headline or chip, and the demo-only controls.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds the error views requirement.

## Impact

- `web/src/routes/+error.svelte`, `web/src/lib/errors.ts`,
  `web/src/hooks.server.ts`, `web/src/hooks.client.ts`, `web/src/error.html`,
  `web/static/404.html`, `web/static/500.html`, `web/src/app.d.ts`, README.

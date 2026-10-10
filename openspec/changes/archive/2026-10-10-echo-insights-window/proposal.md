# Proposal

## Why

The default telemetry window, 7 days, is a constant in the Go settings and
metrics packages and a third copy in the frontend's `telemetryWindow` store,
used when the settings fetch fails. The two sides have to agree and nothing
enforces it; a comment on each side asks whoever changes one to change the
other. Finding 4 of #376: constants that mirror backend config belong behind
the API.

## What Changes

- `GET /api/insights/failures` reports the window it actually used as
  `window`, the requested one or, when it was omitted or unrecognised, the
  server's default.
- The frontend's store has no default. With no saved or cached value it
  sends no window, lets the server choose, and shows the toggle from the
  response.
- The `7d` constant leaves the frontend.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `pipeline-metrics`: the failure insights report the window they cover.

## Impact

- `openapi/openapi.yaml` (`window` on `FailureInsights`, required),
  `internal/metrics/insights.go`, `internal/httpserver/failure_insights.go`.
- `web/src/lib/telemetryWindow.svelte.ts`, `dashboardApi.ts`, and the
  overview and root-cause pages, which show the toggle from the response.

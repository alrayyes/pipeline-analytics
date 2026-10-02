# Proposal

## Why

The dashboard decides what a run's or step's forge status means in the
browser: `failure` and `timed_out` are failures, anything not `completed` is
running, and the runs view uses that to decide whether to keep polling. The
same rule lives in the Go metrics code (`isFailed`, the run-status buckets),
so the two can drift, and a second client such as the Go SDK or the MCP
endpoint has to reimplement it. Finding 1 of #376, from the account rule
that business rules live behind the API.

## What Changes

- Every run and every step the API returns carries an `outcome`: `passed`,
  `failed`, `running`, `queued`, `cancelled`, `skipped` or `unknown`,
  computed by the metrics package from the forge's status and conclusion.
- The frontend maps `outcome` to a badge and a label and no longer
  interprets `status` or `conclusion` strings. The mapping it had is
  deleted in the same change.
- `status` and `conclusion` stay in the response as the forge's raw values.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `pipeline-metrics`: adds a requirement that runs and steps are reported
  with a normalized outcome.

## Impact

- `openapi/openapi.yaml`: `Outcome` schema; `outcome` required on `RunStep`
  and `RunSummary`. Additive for clients that ignore it. The Go, Node, PHP and
  Python SDKs regenerate from it.
- `internal/metrics` (the mapping and its tests), `internal/httpserver` (the
  run and step DTOs).
- `web/`: `statusModel.ts`, `stageProgress.ts`, the runs view and its card,
  the API types; their tests and the runs Playwright spec's fixtures.

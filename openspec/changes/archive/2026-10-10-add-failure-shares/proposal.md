# Proposal

## Why

The root-cause and overview views show each failing step's and each failure
category's percentage of all failures. The API returns counts only, so the
browser would add them up and divide: arithmetic on API data behind a figure
a second client would also want (the account rule that business rules live
behind the API, #376). Tracked in #381.

## What Changes

- `GET /api/insights/failures` adds a `share` to every `stageDistribution`
  entry and a new `categoryBreakdown` (category, occurrences, share), both
  computed over all failed-step occurrences in the window.
- The frontend renders the shares and formats them as percentages.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `pipeline-metrics`: adds a requirement for failure shares.

## Impact

- `openapi/openapi.yaml`: `share` on `StageFailureCount`, a `CategoryCount`
  schema and `categoryBreakdown` on `FailureInsights`, all required.
- `internal/metrics/insights.go`, `internal/httpserver/failure_insights.go`.
- The SDKs regenerate from the spec.

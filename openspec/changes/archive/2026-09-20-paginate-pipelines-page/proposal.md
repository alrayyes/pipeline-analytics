# Proposal

## Why

`GET /api/pipelines` and the Pipelines page (`web/src/routes/+page.svelte`)
fetch and render every distinct `(repo_id, pipeline_name)` pair recorded in
the `runs` table on every load, with no `limit`/`offset` anywhere in the
stack (SQL, `metrics.Store.ListPipelines`, `metrics.Service.ListPipelines`,
the handler, or the frontend fetch). As tracked repos and workflows grow,
that's an unbounded list fetched, summarized (health status, trend), and
rendered in one request — the same problem `GET /api/repos` had before
#141 paginated it. Tracked in GitHub issue #243, grouped under milestone
"Pipelines & Steps pages: pagination" alongside the sibling Steps-page
ticket (#244), which depends on this one landing first since it reuses the
same `metrics.Store.ListPipelines`-derived ordering/pagination plumbing.

## What Changes

- `metrics.Store.ListPipelines` gains `limit`/`offset` parameters and
  returns a `hasMore` flag, fetching `limit + 1` rows and trimming so the
  caller never needs a separate count query — same technique #141 used for
  `ingestion.Store.ListRepos`.
- The underlying SQL (`SELECT DISTINCT repo_id, pipeline_name FROM runs`)
  gains a deterministic `ORDER BY repo_id, pipeline_name` — required for a
  stable page boundary, and a gap that exists independent of pagination
  today (no ordering at all).
- `metrics.Service.ListPipelines` and the `GET /api/pipelines` handler pass
  `limit`/`offset` through and wrap the response as `{pipelines, hasMore}`
  instead of a bare array. **BREAKING**: changes the response shape of
  `GET /api/pipelines` from a bare JSON array to an object.
- The Pipelines page gets a Previous/Next control (mirroring the Repos
  page's), reading `limit`/`offset` from the URL/component state and
  resetting to page 1 whenever the health-status filter, repo selector,
  sort order, or forge filter changes.
- `openapi/openapi.yaml` documents the new query params and response shape
  ahead of the implementation, per `rules/api.md`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: the "Repo/pipeline overview" requirement gains a
  pagination scenario — the overview no longer guarantees every tracked
  pipeline renders in one view when the count exceeds a page.

## Impact

- Backend: `internal/metrics/metrics.go` (`Store` interface,
  `Service.ListPipelines`), `internal/metrics/sqlite/store.go`
  (`ListPipelines` SQL), `internal/httpserver/pipelines.go` (`list`
  handler and a new `pipelineListDTO`).
- API contract: `openapi/openapi.yaml`'s `GET /api/pipelines` operation.
- Frontend: `web/src/routes/+page.svelte` (pagination state, Previous/Next
  control, page reset on filter change).
- Tests: `internal/metrics/metrics_test.go`,
  `internal/metrics/sqlite/store_test.go`,
  `internal/httpserver/pipelines_test.go`, and an e2e pagination scenario
  in `web/tests/e2e/dashboard.spec.ts`.

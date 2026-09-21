# Proposal

## Why

`GET /api/steps/unhealthy` and the Steps page
(`web/src/routes/steps/+page.svelte`) fetch and render every tracked
pipeline's unhealthy-step group in one request, with no `limit`/`offset`
anywhere in the stack: the handler (`unhealthySteps` in
`internal/httpserver/pipelines.go`) calls
`metrics.Service.ListUnhealthySteps`, which calls
`store.ListPipelines(ctx, PipelineListFilter{})` for every tracked
pipeline and aggregates that pipeline's step occurrences into one flat
`[]PipelineStepsGroup`. As tracked repos and workflows grow, that's the
same unbounded-fetch problem `GET /api/pipelines` had before #250
paginated it. Tracked in GitHub issue #244, milestone "Pipelines & Steps
pages: pagination" alongside the Pipelines-page ticket (#243, shipped in
#250), which this one explicitly depends on landing first since it reuses
the same `metrics.Store.ListPipelines`-derived ordering/pagination
plumbing rather than duplicating it.

## What Changes

- `metrics.Service.ListUnhealthySteps` gains `limit`/`offset` parameters
  (via the existing `metrics.PipelineListFilter{Limit, Offset}`, already
  added by #250) and returns a `hasMore` flag, paginating over *pipeline
  groups* -- not individual steps within a group, which stay unbounded
  since a pipeline's own step count is already small and bounded.
- `GET /api/steps/unhealthy` accepts `limit`/`offset` query params and
  wraps its response as `{groups, hasMore}` instead of a bare array.
  **BREAKING**: changes the response shape from a bare JSON array to an
  object.
- The Steps page gains a Previous/Next control mirroring the Pipelines
  page's own (#250), fetching one page of pipeline groups at a time
  instead of the whole tracked-pipeline set.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a new requirement covering the Steps page's
  cross-pipeline unhealthy-steps overview and its pagination -- no
  existing requirement in this capability currently describes that page
  at all (the closest, "Flaky-step callout", only covers visual
  distinction between a flaky and a consistently-failing step, not the
  page's own listing or pagination behavior).

## Impact

- `internal/metrics/metrics.go`: `Service.ListUnhealthySteps` signature
  and its one call site (`internal/httpserver/pipelines.go`'s
  `unhealthySteps` handler).
- `internal/httpserver/pipelines.go`: new `pipelineStepsGroupListDTO`
  wrapper, query-param parsing.
- `openapi/openapi.yaml`: `GET /api/steps/unhealthy` query params and
  response schema.
- `web/src/routes/steps/+page.svelte`: paginated fetch, offset/hasMore
  state, Previous/Next control.
- `web/tests/e2e/dashboard.spec.ts`: Steps-page pagination scenario and
  updated mocks for the wrapped response shape.

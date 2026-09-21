# Tasks

## 1. API contract

- [ ] 1.1 Add `limit`/`offset` query params and a
      `UnhealthyStepsList{groups, hasMore}` response schema to
      `GET /api/steps/unhealthy` in `openapi/openapi.yaml`; verify
      `bunx redocly lint` passes.

## 2. Service

- [ ] 2.1 Change `metrics.Service.ListUnhealthySteps` to
      `ListUnhealthySteps(ctx, window Window, limit, offset int)
      ([]PipelineStepsGroup, bool, error)` in `internal/metrics/metrics.go`:
      keep the existing unfiltered, unpaginated
      `store.ListPipelines(ctx, PipelineListFilter{})` call and per-pipeline
      aggregation, then slice the resulting ordered `[]PipelineStepsGroup`
      by `limit`/`offset` and trim one extra element into `hasMore`, same
      technique `sqlite.Store.ListPipelines` uses for its own `hasMore`;
      verify `internal/metrics/metrics_test.go` covers `limit`/`offset`
      slicing and `hasMore` true/false at the boundary.

## 3. HTTP handler

- [ ] 3.1 Add `pipelineStepsGroupListDTO{Groups
      []pipelineStepsGroupDTO, HasMore bool}` and update
      `pipelinesHandler.unhealthySteps` in `internal/httpserver/pipelines.go`
      to parse `limit`/`offset` query params, call the updated service
      method, and wrap the response; verify
      `internal/httpserver/pipelines_test.go` covers pagination (`hasMore`
      true/false, limit/offset slicing).

## 4. Frontend

- [ ] 4.1 Update `web/src/routes/steps/+page.svelte`'s fetch to read
      `{groups, hasMore}`, add `offset`/`hasMore` state and a `PAGE_SIZE`
      constant matching the Pipelines page's, appending `limit`/`offset`
      to the `/api/steps/unhealthy` request; verify `bun run check` passes
      with no new type errors.
- [ ] 4.2 Add a Previous/Next `Button` pair mirroring the Pipelines
      page's, and reset `offset` to 0 in an `$effect` when the forge
      filter changes (the filter itself stays client-side, applied to the
      current page's fetched groups, per design.md's scoping decision).
- [ ] 4.3 Update the empty-state and grouped-by-repo rendering to read
      from the current page's `groups` rather than an unpaginated whole
      set; verify by reading the rendered output against a fixture with
      more unhealthy pipelines than one page.

## 5. End-to-end coverage

- [ ] 5.1 Add a pagination scenario to `web/tests/e2e/dashboard.spec.ts`
      covering the Steps page: Previous/Next button state at each
      boundary, and offset resetting to 0 when the forge filter changes;
      update every mocked `GET /api/steps/unhealthy` response in that
      journey to the new wrapped shape; verify `bunx playwright test`
      passes.

## 6. Full verification

- [ ] 6.1 Run `go test ./...`, `golangci-lint run ./...`, `bun run
      check`, `bun run lint`, and `bunx playwright test` together; verify
      all pass with no regressions in existing Pipelines/Repos page tests.

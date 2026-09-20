# Tasks

## 1. API contract

- [x] 1.1 Add `limit`/`offset`/`repoId`/`forge` query params and a
      `PipelineList{pipelines, hasMore}` response schema to
      `GET /api/pipelines` in `openapi/openapi.yaml`; verify
      `bunx redocly lint` passes.

## 2. Store and service

- [x] 2.1 Add `metrics.PipelineListFilter{RepoID, Forge, Limit, Offset}`
      and change `metrics.Store.ListPipelines` to
      `ListPipelines(ctx, PipelineListFilter) ([]PipelineRef, bool, error)`
      in `internal/metrics/metrics.go`; verify `go build ./...` succeeds.
- [x] 2.2 Update `internal/metrics/sqlite/store.go`'s `ListPipelines`: add
      `ORDER BY r.repo_id, r.pipeline_name`, an optional
      `JOIN repos p ON p.id = r.repo_id` + `WHERE p.forge = ?` when
      `Forge` is set, a `WHERE r.repo_id = ?` when `RepoID` is set, and
      `LIMIT ?+1 OFFSET ?` when `Limit > 0`, trimming the extra row into
      `hasMore`; verify `internal/metrics/sqlite/store_test.go` covers
      ordering, the forge/repo filters, and the limit+1-row trim.
- [x] 2.3 Update `metrics.Service.ListPipelines` to accept
      `PipelineListFilter` and pass it through, returning `(pipelines,
      hasMore, error)`; update its one other call site
      (`ListUnhealthySteps`, passing `PipelineListFilter{}` to preserve
      today's full-scan behavior) to match the new `Store` signature;
      verify `internal/metrics/metrics_test.go` passes.

## 3. HTTP handler

- [x] 3.1 Add `pipelineListDTO{Pipelines []pipelineSummaryDTO, HasMore
      bool}` and update `pipelinesHandler.list` in
      `internal/httpserver/pipelines.go` to parse `limit`/`offset`/
      `repoId`/`forge` query params, call the updated service method, and
      wrap the response; verify
      `internal/httpserver/pipelines_test.go` covers pagination
      (`hasMore` true/false, limit/offset slicing) and the repo/forge
      filters.

## 4. Frontend

- [x] 4.1 Update `web/src/routes/+page.svelte`'s `PipelineSummary[]`
      fetch to read `{pipelines, hasMore}`, add `offset`/`hasMore` state,
      a `PAGE_SIZE` constant, and a Previous/Next `Button` pair mirroring
      `web/src/routes/repos/+page.svelte`'s; verify `bun run check`
      passes with no new type errors.
- [x] 4.2 Move the forge filter and repo selector into the `/api/pipelines`
      query string (`forge`, `repoId`) instead of filtering the fetched
      array client-side, and reset `offset` to 0 in an `$effect` when
      either changes; verify the existing
      `web/src/lib/pipelinesFilters.svelte.test.ts` still passes and the
      client-side `forgeFilteredPipelines`/`repoFilteredPipelines`
      derivations are removed (health-status filtering and sort stay
      client-side, now applied to the current page only, per design.md's
      scoping decision).
- [x] 4.3 Update the "Showing N of M" summary text to describe the
      current page rather than a whole-set total that pagination no
      longer makes free to compute; verify by reading the rendered text
      against a fixture with more pipelines than one page.

## 5. End-to-end coverage

- [x] 5.1 Add a pagination scenario to `web/tests/e2e/dashboard.spec.ts`
      covering the Pipelines page: Previous/Next button state at each
      boundary, and offset resetting to 0 when the forge filter or repo
      selector changes; verify `bunx playwright test` passes.

## 6. Full verification

- [x] 6.1 Run `go test ./...`, `golangci-lint run ./...`, `bun run
      check`, `bun run lint`, and `bunx playwright test` together; verify
      all pass with no regressions in existing Repos/Steps page tests
      (both call `GET /api/pipelines`-adjacent code paths indirectly via
      shared components).

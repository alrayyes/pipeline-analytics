# Design

## Context

See proposal.md - Why. `metrics.Store.ListPipelines(ctx) ([]PipelineRef,
error)` backs `GET /api/pipelines` and returns every distinct
`(repo_id, pipeline_name)` pair via an unordered `SELECT DISTINCT`. The
same interface method also backs `metrics.Service.ListUnhealthySteps`
(`GET /api/steps/unhealthy`, issue #244), so its signature change affects
both call sites even though only this change (#243) adds pagination
behavior to the Pipelines page itself.

`web/src/routes/+page.svelte` today does a single unpaginated fetch, then
applies the forge filter, repo selector, health-status filter, and sort
order entirely client-side over the full result before grouping by repo
for display.

## Goals / Non-Goals

**Goals:**
- `GET /api/pipelines` paginates via `limit`/`offset`, ordered
  deterministically.
- The repo selector and forge filter compose correctly with pagination
  (moved server-side, same treatment #141 gave the Repos page's forge
  filter).
- The Pipelines page gets a Previous/Next control, resetting to page 1 on
  any filter/sort change.

**Non-Goals:**
- Correct (whole-set) health-status filtering or "most recently run"
  sorting across a paginated result. Per the scoping decision below, both
  become page-local. Making them whole-set-correct would mean computing
  health status (failure rate, duration regression, flaky-step signal)
  and last-run time in SQL/Go ahead of pagination -- a genuine aggregation
  rewrite, tracked as a follow-up rather than folded into this change.
- Pagination for `GET /api/steps/unhealthy` -- that's issue #244, a
  sibling change that reuses this one's `ORDER BY`/pagination plumbing on
  `metrics.Store.ListPipelines`.

## Decisions

- **Filter/pagination scoping** (asked of and confirmed by the user): ship
  plain `limit`/`offset` pagination now, without rearchitecting health
  computation into SQL. The repo selector and forge filter are simple
  attribute matches (`runs.repo_id`, `repos.forge` via a join) and move
  server-side so they don't break pagination's page-size/`hasMore`
  invariants the way an un-pushed-down filter would (the exact bug #141's
  PR description flags: "filtering in Go after paginating in SQL would
  make a page's size and `hasMore` both wrong"). The health-status filter
  and "most recently run" sort stay client-side, now applied per-page
  rather than across the full set -- an accepted, documented temporary
  regression versus today's behavior (see the new dashboard-ui scenario
  "Health-status filter and \"most recently run\" sort apply per page").

- **`Store.ListPipelines` signature**: change from `ListPipelines(ctx)
  ([]PipelineRef, error)` to `ListPipelines(ctx, PipelineListFilter)
  ([]PipelineRef, bool, error)`, where:

  ```go
  type PipelineListFilter struct {
      RepoID string // exact match on runs.repo_id; "" = no filter
      Forge  string // exact match on repos.forge via join; "" = no filter
      Limit  int    // 0 = no limit (used by ListUnhealthySteps, #244's own
                     // pagination lands separately)
      Offset int
  }
  ```

  Mirrors `ingestion.RepoListFilter` (#141) rather than inventing a
  different shape for the sibling pagination feature. `ListUnhealthySteps`
  (existing, unpaginated call site) passes `PipelineListFilter{}` -- zero
  value, no limit, no filter -- preserving its current full-scan behavior
  until #244 adds its own pagination on top.

- **SQL**: add `ORDER BY r.repo_id, r.pipeline_name` (missing today,
  needed for a stable page boundary regardless of pagination) and a
  `JOIN repos p ON p.id = r.repo_id` only when `Forge` is set, since the
  `runs` table has no forge column of its own. `Limit > 0` fetches
  `Limit + 1` rows and the caller trims the last one to compute `hasMore`,
  same technique as #141 -- avoids a separate `COUNT(*)` query.

- **Response shape**: `GET /api/pipelines` wraps its array as
  `{"pipelines": [...], "hasMore": bool}` (**BREAKING** -- was a bare
  array). Named `pipelines`, not `items` or `repos`, matching each
  endpoint's own resource name the way `repoListDTO{Repos, HasMore}` does
  for `GET /api/repos`.

- **Frontend pagination state**: `offset`/`hasMore` as component state in
  `+page.svelte`, `PAGE_SIZE` constant matching the Repos page's, a
  Previous/Next `Button` pair disabled at the respective boundary. An
  `$effect` resets `offset` to 0 whenever the forge filter, repo selector
  (now server-side) changes -- the health filter and sort order don't
  need this since they no longer trigger a re-fetch, only a client-side
  re-filter of the current page.

## Risks / Trade-offs

- [Health-status filter and "most recently run" sort become page-local,
  which can show an empty or partial page even when matching pipelines
  exist elsewhere] → Documented in the spec delta and the issue; a
  `role="status"` hint text already present ("Showing N of M") continues
  to reflect the current page's counts, so the UI doesn't claim a
  whole-set total it can no longer compute cheaply. Revisit if this proves
  confusing in practice -- tracked as a follow-up, not silently accepted
  forever.
- [Adding a `repos` join to `ListPipelines` only when `Forge` is set keeps
  the common (unfiltered) query simple, but means two SQL statement
  shapes to maintain] → Small, well-scoped increase; the same pattern
  already exists implicitly in `ingestion.sqlite.Store.ListRepos`'s
  conditional `WHERE`.

## Migration Plan

No data migration. Deploy is the usual single-binary release; the
response-shape break is caught by `bun run check` against the regenerated
API types and the e2e suite, not by any runtime compatibility shim --
there's no external API consumer to version against (`rules/api.md`
doesn't ask for one on a solo-dev dashboard with no third-party clients).

# Design

## Context

See proposal.md - Why. `metrics.Service.ListUnhealthySteps` (today,
`internal/metrics/metrics.go`) calls `store.ListPipelines(ctx,
PipelineListFilter{})` -- zero value, unfiltered, unpaginated -- to get
every tracked pipeline ref in `(repo_id, pipeline_name)` order, then for
each ref calls `store.PipelineSteps` and `aggregateSteps` to compute
whether that pipeline has a flaky or failing step. A pipeline with none
contributes nothing to the result; the ordered list this builds is what
the handler serializes.

Unlike `GET /api/pipelines`, "has an unhealthy step" is not a column or a
simple `WHERE` clause -- it only exists after `aggregateSteps` runs per
pipeline. `PipelineListFilter{Limit, Offset}` (added by #250) pages over
the *candidate* pipeline set at the SQL layer, before that computation
happens. Passing it straight through here would paginate the wrong set:
a page could come back with zero groups on it even though later,
unfetched pipelines are unhealthy, and `hasMore` would answer "are there
more pipelines" rather than "are there more pipeline groups with an
unhealthy step" -- the same class of bug #250's design doc flagged for
an unpushed-down filter ("filtering in Go after paginating in SQL would
make a page's size and `hasMore` both wrong"), just one layer further in.

## Goals / Non-Goals

**Goals:**

- `GET /api/steps/unhealthy` paginates over pipeline *groups* --
  `limit`/`offset`, a `{groups, hasMore}` response -- so the page and its
  DOM stay bounded as the number of unhealthy pipelines grows.
- The Steps page gets a Previous/Next control, resetting to page 1 when
  the forge filter changes.
- Page ordering is deterministic and matches the Pipelines page's own
  (`repo_id, pipeline_name`), reusing `ListPipelines`'s existing
  `ORDER BY` rather than adding a second one.

**Non-Goals:**

- Reducing the per-load computation itself. Every tracked pipeline's
  step aggregation still runs on every load, same as today -- "has an
  unhealthy step" isn't knowable without it, and pushing that
  computation into SQL (a materialized health signal, computed and kept
  fresh independent of a request) is a genuine aggregation rewrite in
  its own right, the same class of scope #250 declined for health-status
  filtering on the Pipelines page. This change bounds the *response* and
  the *rendered page*, not the backend work computing it.
- Pagination for `GET /api/pipelines` -- that's #243/#250, already
  shipped; this change only reuses its ordering.

## Decisions

- **Paginate the computed result, not the candidate fetch.**
  `ListUnhealthySteps` keeps calling `store.ListPipelines(ctx,
  PipelineListFilter{})` unfiltered and unpaginated (`Limit: 0`) to get
  every candidate ref in stable `(repo_id, pipeline_name)` order --
  reusing that ordering guarantee is what #244's DoD asks for, not a
  second `ORDER BY`. It still computes every pipeline's unhealthy steps
  in Go, same as today, builds the full ordered `[]PipelineStepsGroup`,
  and *then* slices that slice by `limit`/`offset`, trimming an extra
  fetched element into `hasMore` the same way the SQL layer does for
  `ListPipelines` -- consistent behavior, computed one layer higher
  because the filter it's paginating over is Go-side, not a `WHERE`
  clause.
- **Signature**: `ListUnhealthySteps(ctx, window Window, limit, offset
  int) ([]PipelineStepsGroup, bool, error)`. Plain ints rather than
  reusing `PipelineListFilter` -- that struct's `RepoID`/`Forge` fields
  don't apply here (the Steps page's forge filter stays client-side,
  same as today, since it filters by repo membership over an
  already-small fetched page rather than needing a server-side join),
  and reusing it with those fields always zero would suggest a
  server-side repo/forge filter exists when it doesn't.
- **Response shape**: `GET /api/steps/unhealthy` wraps its array as
  `{"groups": [...], "hasMore": bool}` (**BREAKING** -- was a bare
  array). Named `groups`, matching `PipelineStepsGroup`'s own name the
  way `pipelines`/`repos` match their DTOs on the sibling endpoints.
- **Frontend pagination state**: `offset`/`hasMore` component state in
  `+page.svelte`, the same `PAGE_SIZE` constant and Previous/Next
  `Button` pair pattern as the Pipelines page. An `$effect` resets
  `offset` to 0 when the forge filter changes. The forge filter itself
  stays client-side (unlike the Pipelines page's, which moved
  server-side in #250) since the Steps page's candidate set for a filter
  change is still the current page's `groups`, filtered against
  `repos` the page already fetches separately -- there's no `repos.forge`
  join to add here because pagination happens after aggregation, not in
  SQL. This does mean a page can show fewer results than its size, or
  none, once a forge filter narrows it -- documented as accepted
  page-local behavior, same as #250's health-status/sort scoping
  decision.

## Risks / Trade-offs

- [Forge filter is page-local, so a page can show fewer results than its
  size, or none, even when a matching pipeline exists on another page] →
  Documented in the spec delta's "Forge filter change resets to the
  first page" scenario; matches the accepted precedent from #250's own
  health-status/sort scoping decision, not a new regression this change
  introduces.
- [Every tracked pipeline's step aggregation still runs on every load,
  so the pagination doesn't reduce backend work, only response/DOM
  size] → Documented as a Non-Goal above; the ticket's underlying "stays
  fast as pipelines grow" goal for computation cost, not just render
  cost, would need a materialized health signal -- worth its own future
  issue if load time becomes a real problem, not folded into this
  change.

## Migration Plan

No data migration. Deploy is the usual single-binary release; the
response-shape break is caught by `bun run check` against the
regenerated API types and the e2e suite, same as #250.

# Design

## Context

See `proposal.md` - Why. Relevant existing state:

- `ListRecentRuns` (`internal/ingestion/github/client.go`) fetches the
  runs list via `conditionalGet` (ETag-aware), then calls
  `listWorkflowJobs` for every run in the response, unconditionally
  (`conditionalGet(ctx, jobsURL, token, "", &payload)` - the
  `ifNoneMatch` argument is a hardcoded empty string).
- `ingestion.RunSnapshot`/`JobSnapshot` (the client's return types)
  carry no reference to what's already stored - the client has no
  access to `Store`/`RunStore` today, only `ForgeClient`'s own
  interface methods.
- `Reconciler.ReconcileRepo` (`internal/ingestion/reconcile.go`) is
  what actually has both the forge client and the store in scope - it
  calls `client.ListRecentRuns` then loops over the results calling
  `storeRunSnapshot`.

## Goals / Non-Goals

**Goals:**

- No jobs request for a run whose stored state already matches what
  the fresh runs listing reports.
- No change to what ends up stored for any run whose state actually
  changed, including a re-run flipping a completed run back to
  in-progress under the same run ID.

**Non-Goals:**

- Forgejo. Its `ListRecentRuns` already fetches unconditionally by
  design (no rate limit to protect against, no documented conditional-
  request support on that endpoint - see the existing doc comment in
  `internal/ingestion/forgejo/client.go`). Nothing here touches it.
- Adaptive rate-limit backoff. A separate, larger change with its own
  trade-offs against the "at least hourly" guarantee - not bundled
  with this one, which has none.

## Decisions

**The skip check lives in `Reconciler.ReconcileRepo`, not inside the
GitHub client.** The client has no store access and shouldn't gain
one just for this - it would blur `ForgeClient`'s boundary (a pure
forge-API adapter) with the domain's own storage concerns. Instead,
`ListRecentRuns`'s signature gains a way to tell it which runs to skip
jobs for: the simplest shape is passing the reconciler's own
already-known run states in, e.g. `ListRunsRequest` gains a
`KnownRuns map[string]RunState` (forge run ID -> stored status/
conclusion), which `client.ListRecentRuns` checks before calling
`listWorkflowJobs` for each run - keeping the store lookup itself in
`Reconciler` (which already has `RunStore`), and the skip *decision*
in the client (which is what actually knows whether it's about to
make the jobs call).

**"Matches" means status `completed` and the same conclusion**, not
just the same status. GitHub's `conclusion` field (`success`,
`failure`, `cancelled`, etc.) is only meaningful once a run is
`completed`, and comparing it too is what correctly forces a refetch
if a completed run's conclusion is ever corrected after the fact
(rare, but the whole point of a stored-state comparison is not
assuming completed data is permanently immutable without checking).

**`Reconciler` builds `KnownRuns` from a single existing query, not a
per-run lookup.** `RunStore` already has (or gains, if it doesn't) a
way to list a repo's stored runs with their status/conclusion in one
call - looking that up once before calling `ListRecentRuns`, rather
than querying per run ID inside the loop, keeps this from trading one
kind of N+1 (API calls) for another (DB queries).

## Risks / Trade-offs

- **A run's conclusion changing after `completed` without a new run
  ID** -> the comparison catches this (Decisions, above) by checking
  conclusion too, not just status; accepted as sufficiently rare that
  no additional signal (e.g. `updated_at`) is worth comparing on top.
- **`ListRunsRequest` gaining a new field is a signature change to a
  port interface** (`ForgeClient.ListRecentRuns`) -> both
  implementations (`github`, `forgejo`) need updating even though
  Forgejo's ignores it; a small, mechanical change, not a design risk.

## Migration Plan

1. Add `KnownRuns` to `ListRunsRequest` and the stored-run-state query
   `Reconciler` uses to populate it.
2. Update `github/client.go`'s `ListRecentRuns` to skip
   `listWorkflowJobs` for a run present in `KnownRuns` with matching
   status and conclusion.
3. Update `forgejo/client.go`'s signature to accept the same request
   shape (ignoring the new field), so `ForgeClient` stays one
   interface both implement identically.

No rollback concern: purely fewer API calls for identical stored
results, no schema or behavior change for anything that actually
changed.

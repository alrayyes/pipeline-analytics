# Proposal

## Why

`ListRecentRuns` (`internal/ingestion/github/client.go`) lists up to 20
recent runs with a conditional, ETag-aware request, but then calls
`listWorkflowJobs` for *every* one of those runs unconditionally — that
call has no conditional-request logic at all. So the moment anything
changes since the last poll (one new run, one status update), the
reconciler re-fetches jobs for all 20 returned runs, including ones
that completed hours or days ago and can never change again. For any
repo with regular activity, that's up to 21 GitHub API calls per poll
instead of the 1-3 actually needed.

## What Changes

- Before fetching a run's jobs, the reconciler checks whether that run
  is already stored as `completed` with the same conclusion the fresh
  listing reports. If so, it skips the jobs call and keeps the stored
  job/step data instead of re-fetching it.
- A run that's new, still in progress, or whose top-level status
  differs from what's stored (including a re-run flipping a completed
  run back to in-progress under the same run ID) is still fetched
  normally — this only skips a call that would return data identical
  to what's already stored.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `forge-ingestion`: "Reconciliation polling" gains a scenario
  extending its existing "costs no rate-limit budget when unchanged"
  guarantee down to the per-run level, not just the per-repo level.

## Impact

- Backend: `internal/ingestion/github/client.go`'s `ListRecentRuns` and
  `internal/ingestion/reconcile.go`'s `ReconcileRepo` need to consult
  already-stored run state before deciding whether to fetch jobs for a
  given run — see `design.md` for exactly where that check lives.
- Forgejo's client isn't affected by this change (Forgejo's own
  `ListRecentRuns` doesn't have GitHub's rate-limit pressure — see
  `internal/ingestion/forgejo/client.go`'s own doc comment on why it
  fetches unconditionally already).

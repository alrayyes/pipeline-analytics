# Tasks

## 1. Store: known-run-state lookup

- [x] 1.1 Add a `RunStore` method returning each stored run's status
      and conclusion for a repo in one query, with a unit test
      covering a repo with a mix of completed and in-progress runs

## 2. Ports and clients

- [x] 2.1 Add `KnownRuns map[string]RunState` to `ListRunsRequest`
      (forge run ID -> stored status/conclusion)
- [x] 2.2 Update `github/client.go`'s `ListRecentRuns` to skip
      `listWorkflowJobs` for a run in `KnownRuns` with matching status
      and conclusion, with unit tests for: a match (skipped), a status
      mismatch (fetched), and a completed-to-in-progress transition on
      the same run ID (fetched)
- [x] 2.3 Update `forgejo/client.go`'s signature to accept the new
      field (ignored), verifying its existing behavior is unchanged --
      turned out to need no code change at all: `ListRunsRequest` is
      already one shared struct both `ForgeClient` implementations
      take, so adding a field to it doesn't touch Forgejo's signature;
      only the "still ignores it" test was new

## 3. Wiring

- [x] 3.1 Update `Reconciler.ReconcileRepo` to build `KnownRuns` from
      the new store lookup before calling `ListRecentRuns`, and verify
      (with a test double counting calls) that a poll against an
      all-unchanged repo makes zero jobs requests -- verified with the
      real `github.Client` against a fake HTTP server, since "zero
      jobs requests" is an HTTP-level guarantee an in-package fake
      `ForgeClient` can't actually demonstrate

## 4. Ship it

- [x] 4.1 Open a pull request with `Closes #206` and verify CI passes

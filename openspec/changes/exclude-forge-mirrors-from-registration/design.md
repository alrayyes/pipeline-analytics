# Design

## Context

See `proposal.md` - Why. Relevant existing state:

- `ForgeClient.ListAccessibleRepos` (`internal/ingestion/ingestion.go:122-125`)
  already excludes archived/fork/mirror repos, but only for the
  Discover picker's listing. `Registrar.Register`
  (`internal/ingestion/registrar.go:33`) never checks this at all -
  it persists whatever identifier it's given.
- `internal/ingestion/forgejo/client.go`'s `ListAccessibleRepos` reads
  `r.Archived`, `r.Fork`, `r.Mirror` off the Gitea SDK's bulk
  `/user/repos` listing (`ListMyRepos`, which decodes into the same
  `Repository` struct the single-repo endpoint would use). Whether
  Forgejo's *server* populates those fields as reliably on the list
  endpoint as on a single-repo fetch is unconfirmed - 42 known mirrors
  got tracked despite this filter existing, which is the concrete
  evidence something in that path isn't catching them.
- Neither `ForgeClient` implementation
  (`internal/ingestion/github`, `internal/ingestion/forgejo`) has a
  single-repo lookup today - both only ever list repos in bulk
  (`ListAccessibleRepos`) or fetch runs (`ListRecentRuns`).

## Goals / Non-Goals

**Goals:**

- No registration path - Discover, manual entry, or a retried request
  - can create a tracked record for an archived, forked, or mirror
  repo.
- The check is authoritative: a single-repo lookup at the moment of
  registration, not reused Discover data that might be stale or
  incompletely populated.

**Non-Goals:**

- Diagnosing *why* the bulk-list endpoint let mirrors through
  Discover's existing filter. Adding an authoritative check at
  registration makes that root cause moot for tracking purposes,
  whether it's a Forgejo API quirk, an SDK decoding gap, or something
  else - not worth the investigation cost to confirm which.
- Untracking the 42 already-registered mirrors. Manual cleanup via the
  existing untrack endpoint, not new product behavior (proposal.md's
  Impact).

## Decisions

**A new `ForgeClient.GetRepo` method, not reusing `ListAccessibleRepos`
for one repo.** Fetching a single repo's metadata is a different,
smaller operation than paginating a bulk listing - both forge SDKs
(`code.gitea.io/sdk/gitea`, `google/go-github`) already expose a
single-repo GET call directly. Reusing the bulk listing and filtering
client-side for one identifier would mean an unbounded paginated fetch
just to check one repo.

**The check runs against the token being registered, not a
previously-fetched Discover result.** `Registrar.Register` calls
`client.GetRepo` itself, right before `store.CreateRepo`, using the
token from the current request - this is what makes the check
authoritative regardless of how the identifier arrived (Discover,
manual entry, or a stale cached list), directly per the spec's "The
check is authoritative regardless of source" scenario.

**Rejection reuses the same response shape as duplicate detection**
(`fix-repo-registration-duplicate-detection`'s `409` + reason
pattern) rather than inventing a second error format - a
sentinel error per reason (`ErrRepoArchived`, `ErrRepoFork`,
`ErrRepoMirror`) mapped the same way.

## Risks / Trade-offs

- **One extra forge API call per registration attempt** -> negligible
  cost: registration is a rare, user-initiated action, not a hot path,
  and this replaces relying on data that turned out to be unreliable.
- **A repo could change status (e.g. get archived) between this check
  and the webhook-creation step that follows it** -> accepted as a
  narrow race not worth guarding further; the existing degrade path
  already handles a webhook-creation failure gracefully, and an
  archived-after-check repo simply won't produce new runs, which
  reconciliation already tolerates (an empty result, not an error).

## Migration Plan

1. Add `GetRepo` to `ForgeClient` and implement it in both
   `internal/ingestion/github` and `internal/ingestion/forgejo`.
2. Add the archived/fork/mirror sentinel errors and the check in
   `Registrar.Register`, before `store.CreateRepo`.
3. Wire the new error cases into the same `409`-style handling
   `fix-repo-registration-duplicate-detection` adds to `register`'s
   handler (shared response shape, not a duplicate code path).

No rollback concern: purely an additional guard, no schema change.

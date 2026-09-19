# Proposal

## Why

Discover's "already tracked, don't offer it again" check only looks at
the currently loaded page of `/api/repos` (`PAGE_SIZE = 20` in
`web/src/routes/repos/+page.svelte`), not every tracked repo. Past 20
tracked repos on a forge, an already-tracked one reappears in Discover
as if it were new. Following it re-submits `POST /api/repos`, which
hits the `repos` table's `UNIQUE (forge, forgejo_instance_url,
identifier)` constraint and fails — but the handler
(`internal/httpserver/repos.go`) turns *any* registration error into
an opaque `500 internal_error "register repo"` with no log line
anywhere in the file, so the failure is silent and unexplained from
both the UI and the server logs.

## What Changes

- The "already tracked" check Discover uses to filter its results
  covers every tracked repo on that forge/instance, not just the
  current page.
- Registering an identifier that's already tracked on the same forge
  (+ instance, for Forgejo) returns a distinct, clearly-worded
  response (not a generic internal error) instead of failing on the
  underlying DB constraint.
- That rejection, and any other registration failure, is logged
  server-side.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `forge-ingestion`: "Repo tracking registration" gains a requirement
  that re-registering an already-tracked repo is rejected clearly
  rather than failing opaquely.

## Impact

- Frontend: `web/src/routes/repos/+page.svelte`'s `trackedIdentifiers`
  needs the full tracked set, not just the current page — either a
  separate unpaginated identifiers fetch, or a bump to the existing
  `/api/repos` call for this specific check.
- Backend: `internal/httpserver/repos.go`'s `register` handler
  distinguishes a duplicate-registration error from other failures
  (e.g. a sentinel error from `Registrar.Register`/`CreateRepo`) and
  returns `409` for it; both that case and any other failure get a
  `slog` call.
- Database: a new migration adds a normalized unique index — see
  design.md's Context/Migration Plan for why the table's existing
  `UNIQUE` constraint turned out not to fire for GitHub repos at all.

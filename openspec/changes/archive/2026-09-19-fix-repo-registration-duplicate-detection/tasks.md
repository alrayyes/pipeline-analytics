# Tasks

## 1. Backend: detect and surface duplicates

- [x] 1.1 Add `ingestion.ErrRepoAlreadyTracked` and make `CreateRepo`
      return it when the insert fails on the `UNIQUE (forge,
      forgejo_instance_url, identifier)` constraint, with a unit test
      covering both the constraint-violation and a genuine other
      insert failure -- plus migration `00005` fixing the constraint
      itself, found not to fire for GitHub rows at all (see
      design.md's Context)
- [x] 1.2 Add `Store.ListRepoIdentifiers(ctx, forge, instanceURL)
      ([]string, error)` (unpaginated), with a unit test covering more
      than one page's worth of repos
- [x] 1.3 Expose the identifiers list to the frontend (new lightweight
      endpoint or param) and verify it returns every tracked
      identifier regardless of count
- [x] 1.4 Update `register`'s handler: `409` + `slog.ErrorContext` for
      `ErrRepoAlreadyTracked`, `500` + `slog.ErrorContext` for
      anything else, and verify both paths log

## 2. Frontend: fix the Discover exclusion

- [x] 2.1 Switch `trackedIdentifiers` to the new identifiers source
      instead of the paginated `repos` list, and verify (with more
      than 20 tracked repos in a test fixture) that none of them
      reappear in Discover
- [x] 2.2 Show the "already tracked" `409` distinctly in
      `handleFollowSelected`'s per-repo result, rather than the
      generic "Could not register." message

## 3. Ship it

- [x] 3.1 Open a pull request with `Closes #197` and verify CI passes

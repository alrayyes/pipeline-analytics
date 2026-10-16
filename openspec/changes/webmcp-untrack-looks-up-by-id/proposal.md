# Proposal

## Why

The `untrack_repo` confirmation names the repository by reading up to 500 from
the list and finding the id in it, so a repository past the first 500 is named
by its bare id. `GET /api/repos/{repoId}` now exists (`get-repo-by-id`), so the
tool can ask for the one repository (#702).

## What changes

- The lookup requests the repository by id. An unknown id, a failed request or
  one slower than three seconds still names the id, so the question is never
  blocked by it.

## Impact

- Added requirement to `webmcp-support`.

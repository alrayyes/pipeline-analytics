# Proposal

## Why

`GET /api/repos/{repoId}` and the MCP `get_repo` tool already tell a missing
repo (404) from a store failure (500) with `ingestion.ErrRepoNotFound`. The
other callers of the repo lookup and removal still read any error as not
found, or wrap it without checking the sentinel, so a database failure looks
like a missing repo there.

## What changes

- The usage endpoint, the DELETE endpoint and the MCP `get_repo_usage` and
  `untrack_repo` tools answer not-found only for `ErrRepoNotFound`. Any other
  error is logged and answered 500 (`internal_error`) or returned as a tool
  error, without its detail.
- The job log and run action services map `ErrRepoNotFound` to the job or run
  not being found, since a repo's runs go with it. Any other lookup error stays
  a 500.
- OpenAPI documents the 500 on each affected operation.
- The webhook receiver's 401 for an unknown repo is deliberate (it must not
  reveal which repos exist) and is left alone.

## Impact

- Modified requirement in `forge-ingestion`.

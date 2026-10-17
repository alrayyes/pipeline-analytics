# Proposal

## Why

`GET /api/repos/{repoId}` answers 404 for any error from the store. The `Store`
port has no not-found sentinel of its own (the SQLite adapter's lives a layer
below), so a database failure reads as "no such repository". The API, the MCP
`get_repo` tool and the untrack confirmation all tell the person the repository
is gone when the database is down (#712).

## What changes

- `ingestion.ErrRepoNotFound` becomes part of the `Store` port. The SQLite
  adapter returns it when no row matches, in place of its own sentinel.
- The lookup answers 404 only for that sentinel. Any other error is logged and
  answered 500 with the API's `internal_error` body, without the cause.
- The OpenAPI spec documents the 500 on `getRepo`.
- The MCP `get_repo` tool reports a store failure as an error that is not a
  not-found.

## Impact

- Added requirement to `forge-ingestion` and `mcp-endpoint`.
- Other store lookups share the pattern and are left for their own tickets.

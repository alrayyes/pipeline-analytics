# Proposal

## Why

The Pipelines page filters by health and sorts by last run in the browser,
over the one page of results it loaded. A pipeline that matches the filter
on a later page never shows, and a page can come back short or empty while a
match exists elsewhere. The spec documents that as a known limit. It is also
a rule living in the front end: which pipelines are unhealthy is decided by
the server, and a second client (the MCP endpoint, an SDK) asking for "the
unhealthy pipelines" gets the same wrong, page-local answer. Finding 2 of
#376.

## What Changes

- `GET /api/pipelines` takes `health` (`healthy` or `unhealthy`) and `sort`
  (`name`, the default, or `lastRun`). Both are applied across every matching
  pipeline before paging, so `hasMore` and page sizes are right.
- The MCP `list_pipelines` tool takes the same two inputs.
- The Pipelines page sends them and drops its own filtering and its
  comparator. An unknown `health` or `sort` is a 400.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: the health filter and the most-recent-run sort apply across
  every tracked pipeline, not the loaded page.

## Impact

- `openapi/openapi.yaml`: two optional query parameters.
- `internal/metrics` (filter, sort, parsing, the whole-set path),
  `internal/httpserver` (the handler and the MCP tool).
- `web/`: the Pipelines page and the Playwright fixtures, which have to honour
  the new parameters the way the server will.

## Non-goals

Counting how many pipelines match each health value, for a badge on the
filter, is not part of this.

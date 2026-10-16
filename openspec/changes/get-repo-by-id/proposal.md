# Proposal

## Why

There is no way to fetch one tracked repository. The WebMCP `untrack_repo` tool
reads up to 500 repositories to find the identifier it names in its question,
and names the bare id for one past the first 500. `rules/mcp.md` and
`rules/frontend.md` put a gap like this in the spec, not in a workaround in the
client (#702).

## What changes

- `GET /api/repos/{repoId}` returns one tracked repository, token masked as in
  the list, and answers 404 with the API's error body for an unknown id.
- A `get_repo` tool on the MCP server and in WebMCP, so the operation has its
  tool in the same change.
- A follow-up change moves `untrack_repo` onto it.

## Impact

- Added requirement to `forge-ingestion`, `mcp-endpoint` and `webmcp-support`.

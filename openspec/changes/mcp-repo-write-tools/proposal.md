# Proposal

## Why

Registering a repo, discovering what a token can reach and untracking a repo
have no MCP tool (#615). `rules/mcp.md` wants every operation reachable through
MCP, with the same authorisation and validation as the endpoint.

## Decisions

- **The endpoints' authorisation, which is not session-only.** These three
  endpoints take any caller the API gate lets in, an API token included, so the
  tools do too. Opening them further than the endpoints would break the rule,
  and closing them tighter would leave a token-holding agent without a
  capability the API gives it.
- **A saved forge token is used the way the endpoint uses it.** With no `token`
  argument, registration and discovery fall back to the token saved for the
  forge and instance, so an agent can register a repo or list what the saved
  token reaches without ever seeing it. The tools never return the token, and
  `register_repo` reports it masked.
- **One implementation.** Registering, discovering and untracking move out of
  the handlers into functions that return a problem (status, code, message), and
  the handlers and the tools both call them, so a conflict, a length limit or a
  missing token reads the same through either.

## What Changes

- `register_repo`, `discover_repos` and `untrack_repo`.
- `discover_repos` returns `{ "identifiers": [...] }`; the endpoint answers a
  bare array, and a tool's structured result must be an object.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: the tools cover repo registration, discovery and removal.

## Impact

- `internal/httpserver/mcp_repo_writes.go`, `repos.go`, `forge_tokens.go`,
  `mcp.go`, their tests.

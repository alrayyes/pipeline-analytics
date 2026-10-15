# Proposal

## Why

Three read operations have no MCP tool: the version, the repo list and a
forge's repo identifiers (#615). `rules/mcp.md` wants every operation reachable
through MCP. This slice adds them, taking them off the parity test's pending
list. The rate-limit and settings reads follow, since they need more wiring.

## What Changes

- `get_version`, `list_repos` and `list_repo_identifiers`, with the inputs the
  endpoints take.
- The list body and the forge check are extracted from the handlers, so the
  tool and the endpoint report a repo the same way (token masked) and reject
  the same forge values.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: the tools cover the service version and the tracked repos.

## Impact

- `internal/httpserver/mcp_repos.go`, `repos.go`, `mcp.go`, their tests.

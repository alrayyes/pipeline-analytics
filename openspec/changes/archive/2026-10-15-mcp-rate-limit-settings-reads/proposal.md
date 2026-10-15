# Proposal

## Why

The GitHub rate-limit read and the settings read have no MCP tool (#615).
`rules/mcp.md` wants every operation reachable through MCP. Both need more than
the stores the earlier tools use: the rate-limit reporter, and for settings the
signed-in account.

## What Changes

- `get_github_rate_limit` reports, per GitHub token, the repos it serves and
  its last observed status, built by the same grouping the endpoint uses. The
  endpoint answers an array; the tool wraps it as `tokens`, since an MCP tool's
  structured result is an object.
- `get_settings` returns the account's settings beside the server's defaults,
  for the account the call authenticated as, whether by session or token.
- `authInfoFrom(ctx)` lets a tool read who is calling, as the handlers do from
  the request.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: the tools cover the rate-limit status and the settings.

## Impact

- `internal/httpserver/mcp_repos.go`, `mcp.go`, `insights.go`, `session.go`,
  their tests.

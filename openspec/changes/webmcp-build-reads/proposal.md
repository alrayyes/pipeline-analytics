# Proposal

## Why

The running version and GitHub rate-limit use have no WebMCP tool, though the
MCP server has `get_version` and `get_github_rate_limit` (the parity work,
`rules/mcp.md`). They are the last two API operations without one. With them,
every operation in `openapi.yaml` is a WebMCP tool or carries a written
`x-agent-excluded` reason, so the parity test's temporary pending list can go.

## Decisions

- **The tools send the footer's and the Insights page's own requests.** Both
  pages built their request inline, so they move to `fetchVersion` and
  `fetchGitHubRateLimit` in `dashboardApi.ts` first (own commit, no behaviour
  change) and the tools call the same functions.
- **Both are plain reads; neither asks.** The rate-limit read shows tokens masked
  to their last four characters, as the page does.
- **The rate-limit list is wrapped as `{ tokens }`**, since a tool result is an
  object, matching the MCP server's result.
- **No new exclusions.** Every operation without a tool already carried an
  `x-agent-excluded` reason in the spec (probes, the MCP endpoint itself, saved
  forge tokens, the sign-in ceremony, API tokens, credentials and the
  webhooks), so the end state needs nothing added.
- **The pending list is deleted**, with the test that kept it honest. A new
  operation with no tool and no exclusion now fails the parity test directly.

## What Changes

- `get_version` and `get_github_rate_limit`.
- `fetchVersion` and `fetchGitHubRateLimit`, used by the footer, the Insights
  page and the tools.
- The parity test loses its pending list.

## Capabilities

### Modified Capabilities

- `webmcp-support`: the tools cover the version and GitHub rate-limit use, and
  the parity test has no pending list.

## Impact

- `web/src/lib/dashboardApi.ts`, `webmcpTools.ts`, `webmcpOperations.ts`, the
  generated schemas, the footer, the Insights page and their tests.

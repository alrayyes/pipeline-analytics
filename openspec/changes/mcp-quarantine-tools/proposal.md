# Proposal

## Why

Quarantining a flaky step and clearing the mark have no MCP tool (#615).
`rules/mcp.md` wants every operation reachable through MCP. The endpoints are
session-only, because a mark changes what counts as healthy; the tools keep
that rule, as the run-action tools do.

## What Changes

- `quarantine_step` (with an optional note) and `unquarantine_step`.
- `requireToolSession` and `mcpToolError` are the shared session check and error
  type; the run-action tools move onto them.
- The endpoints' error mapping is extracted (`classifyQuarantineError`), so a
  note over 500 characters or an unknown pipeline reads the same either way.
- The schema-parity test now also checks a tool against the operation's JSON
  request body, so `note` is covered.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: the tools cover quarantine.

## Impact

- `internal/httpserver/mcp_quarantine.go`, `mcp.go`, `mcp_run_actions.go`,
  `quarantine.go`, their tests.

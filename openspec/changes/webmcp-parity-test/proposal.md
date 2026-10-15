# Proposal

## Why

WebMCP has 3 tools against 44 API operations, and its only test asserts that
count (#615). `rules/mcp.md` wants every operation reachable through WebMCP,
with exclusions written down. This slice adds the test that makes the gap
visible, the WebMCP twin of the MCP parity test.

## What Changes

- `webmcpParity.test.ts` reads every operation in `openapi.yaml` and wants a
  WebMCP tool, the `x-agent-excluded` reason the MCP test already honours, or
  a place on a temporary pending list.
- `pendingOperations` names the 22 operations still to get a tool. Each
  follow-up pull request shrinks it and the last one deletes it.

## Capabilities

### Modified Capabilities

- `webmcp-support`: a requirement that every operation has a tool or a
  written exclusion.

## Impact

- `web/src/lib/webmcpParity.test.ts` only.

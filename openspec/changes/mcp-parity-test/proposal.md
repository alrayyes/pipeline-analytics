# Proposal

## Why

The MCP parity test only reads GET paths, so a write operation can ship with
no tool and nothing fails (#615). `rules/mcp.md` wants every operation in
`openapi.yaml` reachable through MCP and WebMCP, with exclusions rare and
written down. This is the first slice: the test that makes the gap visible.
Tools follow in one pull request per resource.

## What Changes

- The parity test reads every operation ID, any method, and wants a tool, an
  `x-agent-excluded` reason on the operation, or a place on a temporary
  pending list.
- Sign-in, token, passkey and forge-token operations, the webhook receivers,
  the probes and `/api/mcp` itself get `x-agent-excluded` with a reason.
- `pendingMCPTools` names the 13 operations still to get a tool. Each
  follow-up pull request shrinks it and the last one deletes it.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: the coverage requirement now spans every operation, not
  only reads.

## Impact

- `openapi/openapi.yaml` (an extension on 19 operations),
  `internal/httpserver/mcp_spec_parity_test.go`.

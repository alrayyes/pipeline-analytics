# Proposal

## Why

Re-run and cancel have no WebMCP tool, though the MCP server has both (the
parity work, `rules/mcp.md`). An in-browser agent sharing the page with its user
cannot do what the Runs page buttons do. These are the first WebMCP tools that
write, so this slice also settles how a WebMCP write is confirmed.

## Decisions

- **The tools call the page's own `rerunRun` and `cancelRun`.** They send the
  same request as the buttons, as the signed-in user, so the endpoint applies
  its own session rule and error mapping. The refusal text is `actionMessage`,
  the sentence the page shows.
- **A person says yes first.** The Runs page asks before it cancels; a tool has
  no click to give. Each call asks through the agent's `requestUserInteraction`
  when the browser offers one, and a native `confirm()` otherwise. A page with
  neither answers no. Declined means nothing is sent. Rejected: confirming only
  the cancel, since a re-run spends runner minutes on someone else's system;
  trusting the agent host to ask, which would make the safety depend on a client
  we do not control.
- **Both are annotated `destructiveHint`**, so a host can gate them as well.

## What Changes

- `rerun_run` and `cancel_run`, each taking a `runId`.
- The "No write tool" requirement is replaced by one that lets a tool write
  only after a person confirms.
- Both come off the parity test's pending list.

## Capabilities

### Modified Capabilities

- `webmcp-support`: write tools exist, behind a confirmation.

## Impact

- `web/src/lib/webmcpTools.ts`, `webmcp.d.ts`, `webmcpOperations.ts`, the
  generated schemas and their tests.

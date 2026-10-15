# Proposal

## Why

Re-run and cancel have no MCP tool (#615). `rules/mcp.md` wants every
operation reachable through MCP, and the earlier `forge-run-actions` decision
kept both session-only because they write to someone else's system. This slice
adds the tools and keeps that rule: they refuse an API token exactly as the
endpoints do.

## Decisions

- **Same authorisation as the endpoint.** The tools and the handlers share the
  error mapping (`classifyRunActionError`) and the session-required message,
  and the tool checks `ViaSession` before asking the forge. A caller with only
  an API token gets `unauthenticated`; a caller with a session, such as the
  in-browser agent, can act. Rejected: opening them to API tokens, which would
  let an agent holding a token re-run or cancel pipelines on the forges;
  excluding them with `x-agent-excluded`, which would leave the in-browser
  agent unable to do what the signed-in user can.
- **The result says the forge took the request.** `accepted` is true when the
  forge answered, not when the run has changed.

## What Changes

- `rerun_run` and `cancel_run`, each taking a `runId`.
- The endpoint's error mapping is extracted and shared.
- The "No write tool" requirement is replaced: tools may write, but only for a
  session, and the first two do.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: the read-only rule becomes a session-only rule for write
  tools. The spec's Purpose still calls the server read-only and is rewritten
  when the last write tool lands and this is archived.

## Impact

- `internal/httpserver/mcp_run_actions.go`, `run_actions.go`, `mcp.go`, their
  tests; `openspec/changes/forge-run-actions/proposal.md` (a note).

# Proposal

## Why

Quarantining a flaky step has no WebMCP tool, though the MCP server has both
`quarantine_step` and `unquarantine_step` (the parity work, `rules/mcp.md`). An
in-browser agent that finds a flaky step cannot mark it as the Flaky page can.

## Decisions

- **The tools call the page's own `quarantineStep` and `unquarantineStep`.**
  The request, the step escaping and the session rule are the page's and the
  endpoint's. A refusal surfaces as the status the server answered.
- **No confirmation prompt.** The Flaky page marks a step with one click and no
  dialog. The mark lasts 30 days, is recorded in this service only and contacts
  no forge, and clearing it undoes it. A prompt here would be stricter than the
  page for a reversible local change. Re-run, cancel and untrack, which act on a
  forge or delete data, do ask.
- **The note is an input like any other.** The body property comes from the
  spec's `QuarantineRequest` through the schema generator, which now reads JSON
  request bodies.

## What Changes

- `quarantine_step` and `unquarantine_step`.
- The schema generator merges a JSON request body's properties into the tool
  schema.
- Both operations come off the parity test's pending list.

## Capabilities

### Modified Capabilities

- `webmcp-support`: the tools cover quarantining a flaky step.

## Impact

- `web/src/lib/webmcpTools.ts`, `webmcpOperations.ts`, the generated schemas,
  `web/scripts/generate-webmcp-schemas.ts` and their tests.

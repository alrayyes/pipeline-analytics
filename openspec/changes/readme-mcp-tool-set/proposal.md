# Proposal

## Why

The README's MCP section says the server has eleven read-only tools. It
registers twenty-six, and several act: they re-run or cancel a run, quarantine
a step, change settings, and register or untrack a repository. A hard-coded
count goes stale with every tool, and "read-only" was never kept true.

## What changes

- The README names the tool groups and says which tools write, and which need a
  session rather than an API token, without a count. The parity test already
  fails when an operation has neither a tool nor a written exclusion, so the
  list can't drift silently.
- CONTRIBUTING gains a paste-able `govulncheck ./...` line next to the note that
  the check has no hook.

## Impact

- Added requirement to `mcp-endpoint`.

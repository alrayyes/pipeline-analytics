# Proposal

## Why

The MCP endpoint mirrored six REST operations when it shipped. The telemetry
work since then added the failure insights, the run list and a run's steps,
and none reached MCP, so an agent could be told about a failed run by
`list_pipeline_flaky_runs` and have no tool to open it. Tracked in #401.

## What Changes

- Three tools: `get_failure_insights`, `list_runs` and `get_run_steps`, the
  same translation the REST handlers do.
- A test that fails when a GET operation in `openapi.yaml` has neither a
  tool nor a reasoned entry on an exclusion list.
- `README.md` says nine tools and where the rule is enforced.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: tool coverage follows the read API.

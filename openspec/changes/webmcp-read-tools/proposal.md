# Proposal

## Why

WebMCP covers 3 of the 25 operations an agent may reach, and the MCP server
already has 12 reads (#615). `rules/mcp.md` wants every operation reachable
through both. This slice adds the six reads whose fetch function the dashboard
already has, so each tool calls the same code the page does.

## What Changes

- `list_pipeline_steps`, `get_job_log`, `get_failure_insights`, `list_runs`,
  `list_flaky_steps` and `list_branches`, named as the MCP tools are.
- `list_pipeline_steps` takes the `window` and `get_job_log` the `lines` the
  API accepts, so no parameter of the operation is out of an agent's reach;
  `fetchPipelineSteps` and `fetchJobLog` gain a params argument for them.
- The six come off the parity test's pending list.

Not here: `get_run_steps`, `list_pipeline_flaky_runs` and
`list_unhealthy_steps`. Their pages call `fetch` inline, so each needs a shared
fetch function first and is its own slice.

## Capabilities

### Modified Capabilities

- `webmcp-support`: the tools cover runs, logs, flaky steps, insights and
  branches as well.

## Impact

- `web/src/lib/webmcpTools.ts` and its tests.

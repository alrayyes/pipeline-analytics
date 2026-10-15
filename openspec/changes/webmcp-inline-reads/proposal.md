# Proposal

## Why

Three read operations have no WebMCP tool because no shared fetch function
exists for them (#615): a run's steps and a step's flaky runs are fetched inline
by their pages, and the unhealthy-steps list is API-only, with no page.
`webmcpTools.ts` promises that a tool calls the function its page calls, so the
tools need those functions to exist first.

## What Changes

- `fetchRunSteps`, `fetchFlakyRuns` and `fetchUnhealthySteps` in
  `dashboardApi.ts`, escaping their arguments and throwing an `ApiError` that
  carries the status.
- The run page and the flaky-runs page use them instead of their own `fetch`,
  keeping their messages: a 404 is "not found", another failure "could not
  load", and a network or parse failure "could not reach the server".
- Component tests for both pages (found, empty, 404, a failed load, an
  unreachable server), since nothing tested those branches at any layer before
  they were rewritten.
- `get_run_steps`, `list_pipeline_flaky_runs` and `list_unhealthy_steps`, named
  as the MCP tools are.

## Capabilities

### Modified Capabilities

- `webmcp-support`: the tools cover a run's steps, a step's flaky runs and the
  unhealthy steps.

## Impact

- `web/src/lib/dashboardApi.ts`, `webmcpTools.ts`, the run and flaky-runs pages,
  their tests.

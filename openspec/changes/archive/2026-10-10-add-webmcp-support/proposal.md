# Proposal

## Why

pipeline-analytics' dashboard (`web/`) is a page a human reads — the
pipeline overview, a pipeline's trend charts, and its usage view are all
rendered for eyes, with nothing exposed to an in-browser agent working
alongside the person using it. Tracked as
[alrayyes/pipeline-analytics#296](https://github.com/alrayyes/pipeline-analytics/issues/296),
queued behind #295 (the backend MCP endpoint, already merged), which
serves the same kind of data to a *remote* MCP client — this is the
browser-side counterpart for an agent already sharing the page.

## What Changes

- Three read-only [WebMCP](https://github.com/webmachinelearning/webmcp)
  tools, registered via `document.modelContext.registerTool()`:
  `list_pipelines` (the overview), `get_pipeline` (a pipeline's health
  and trend data), and `get_repo_usage` (a repo's runner-minutes usage)
  — covering the three views the issue's approach section names
  (overview, per-pipeline detail, usage) and satisfying its "at least
  the overview and per-pipeline detail views" acceptance criterion.
- A new `web/src/lib/dashboardApi.ts`: the three endpoints' fetch calls,
  pulled out of `+page.svelte`'s inline `fetch('/api/...')` calls into
  one function each, so the page and the new WebMCP tool call the same
  code rather than the tool re-implementing its own fetch — the "second
  path to the same API" the issue's approach section explicitly warns
  against. The three pages (`web/src/routes/+page.svelte`,
  `web/src/routes/pipelines/[id]/+page.svelte`,
  `web/src/routes/repos/[id]/usage/+page.svelte`) call the extracted
  functions instead of fetching inline; their own component-local state
  and error handling stay where they are.
- Tool registration lives in `web/src/routes/+layout.svelte`, guarded by
  `if (document.modelContext)` — this is a pre-stable, actively-changing
  browser API (draft spec, experimental in Chrome as of writing, no
  stable release yet), so registration has to be a no-op everywhere else
  rather than a hard dependency.
- No write tool: every view these tools cover is already read-only, so
  this isn't a design choice so much as there being nothing to leave
  out — same as the backend MCP endpoint.
- CONTRIBUTING.md's Frontend section gets a line on what's registered
  and where, next to the existing build instructions.

## Capabilities

### New Capabilities

- `webmcp-support`: WebMCP tool declarations on the dashboard exposing
  its already-fetched pipeline overview, pipeline detail/trend, and
  repo usage data to an in-browser agent.

### Modified Capabilities

(none — `dashboard-ui`'s existing requirements describe what's
*presented*, which doesn't change; this exposes the same data through a
second surface without altering what the dashboard shows or how.)

## Impact

- New `web/src/lib/dashboardApi.ts` (extracted fetch functions).
- New `web/src/lib/webmcpTools.ts` (or similar — tool definitions,
  decided in design.md).
- `web/src/routes/+layout.svelte` (registers the tools on mount).
- `web/src/routes/+page.svelte`,
  `web/src/routes/pipelines/[id]/+page.svelte`,
  `web/src/routes/repos/[id]/usage/+page.svelte` (call the extracted
  functions instead of fetching inline).
- `CONTRIBUTING.md` (Frontend section).
- No backend changes — this is frontend-only; the dashboard already
  authenticates every `/api/...` request via its session cookie, so a
  WebMCP tool running in the same authenticated page inherits that for
  free.

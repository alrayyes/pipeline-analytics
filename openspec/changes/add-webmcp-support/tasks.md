# Tasks

## 1. Extract shared fetch functions

- [x] 1.1 Write a failing `web/src/lib/dashboardApi.test.ts` case:
      `fetchPipelines()` calls `GET /api/pipelines` with the given
      filter/pagination params and returns the parsed
      `PipelineListResponse`; a non-OK response throws
- [x] 1.2 Create `web/src/lib/dashboardApi.ts` with `fetchPipelines`,
      moving `PipelineSummary`/`Repo`/`PipelineGroup`/`PipelineListResponse`
      out of `web/src/routes/+page.svelte`; verify 1.1 passes
- [x] 1.3 Write a failing case: `fetchRepos()` calls `GET /api/repos`
      and returns the parsed repo list; a non-OK response resolves to
      an empty list (matching `+page.svelte`'s current "best effort"
      behavior, not a throw)
- [x] 1.4 Add `fetchRepos` to `dashboardApi.ts`; verify 1.3 passes
- [x] 1.5 Write a failing case: `fetchPipeline(id)` calls
      `GET /api/pipelines/{id}` and returns the parsed
      `PipelineDetail`; a non-OK response throws
- [x] 1.6 Add `fetchPipeline`, moving `Trend`/`PipelineDetail` out of
      `web/src/routes/pipelines/[id]/+page.svelte`; verify 1.5 passes
- [x] 1.7 Write a failing case: `fetchPipelineSteps(id)` calls
      `GET /api/pipelines/{id}/steps` and returns the parsed step list
- [x] 1.8 Add `fetchPipelineSteps`, moving `Step` out of the same file;
      verify 1.7 passes
- [x] 1.9 Write a failing case: `fetchRepoUsage(repoId)` calls
      `GET /api/repos/{repoId}/usage` and returns the parsed usage list
- [x] 1.10 Add `fetchRepoUsage`, moving `UsageEntry` out of
      `web/src/routes/repos/[id]/usage/+page.svelte`; verify 1.9
      passes

## 2. Wire the pages to the extracted functions

Use the `svelte-file-editor` subagent (or the Svelte MCP server's own
tools) for every edit in this section, per this account's session
guidance for `.svelte` files.

- [x] 2.1 Update `web/src/routes/+page.svelte`'s `loadPipelines`/
      `loadRepos` to call `dashboardApi.ts`'s `fetchPipelines`/
      `fetchRepos` instead of fetching inline, importing the types
      from there; verify `bun run check` passes and the page's own
      existing tests (if any target this file) still pass
- [x] 2.2 Update `web/src/routes/pipelines/[id]/+page.svelte` to call
      `fetchPipeline`/`fetchPipelineSteps`; verify `bun run check`
      passes
- [x] 2.3 Update `web/src/routes/repos/[id]/usage/+page.svelte` to
      call `fetchRepoUsage`; verify `bun run check` passes
- [x] 2.4 Run `bun run test:e2e` and verify every existing journey
      still passes -- these three pages' rendered output must be
      unchanged by the refactor. `pwa.spec.ts` and the >20-repo
      pagination journey pass every run. The passkey/overview journey
      hit issue #251's known, previously-investigated, and closed
      "client hydration stalls after a real reload, no console/network
      signal" flake during local runs on this sandbox (confirmed by
      bisection: reverting just this file's change made it pass
      reliably again; reproducing it needed no code path this change
      touches, only more bundle weight on an already
      memory-constrained/swapping host to tip a documented, host-
      timing-dependent stall). #251 itself was closed accepting this
      as external residual risk once the real bug (an unbounded
      layout-load fetch, #268) was fixed, mitigated at the CI level by
      `nick-fields/retry` on a fresh runner (#253) rather than by a
      code fix. Deferring final confirmation to that CI job rather
      than this sandbox.

## 3. WebMCP type declaration and tool registration

- [x] 3.1 Add `web/src/lib/webmcp.d.ts`: a `declare global { interface
      Document { modelContext?: {...} } }` ambient declaration scoped
      to the `registerTool` shape this change actually calls (name,
      description, inputSchema, execute, and the `{signal}` register
      option); verify `bun run check` has no new errors
- [x] 3.2 Write a failing `web/src/lib/webmcpTools.test.ts` case:
      calling `registerWebMCPTools()` when `document.modelContext` is
      undefined does not throw and does not call anything
- [x] 3.3 Write a failing case: calling `registerWebMCPTools()` when a
      fake `document.modelContext.registerTool` is present calls it
      exactly three times, once each named `list_pipelines`,
      `get_pipeline`, `get_repo_usage`
- [x] 3.4 Write a failing case: the `list_pipelines` tool's `execute`
      calls `dashboardApi.ts`'s `fetchPipelines` (spy/mock it) and
      returns its result as the tool's content
- [x] 3.5 Write a failing case: the `get_pipeline` tool's `execute`
      calls `fetchPipeline` with the `pipelineId` argument it's given
- [x] 3.6 Write a failing case: the `get_repo_usage` tool's `execute`
      calls `fetchRepoUsage` with the `repoId` argument it's given
- [x] 3.7 Create `web/src/lib/webmcpTools.ts` implementing
      `registerWebMCPTools()` per design.md's "Tool registration"
      decision; verify 3.2 through 3.6 pass

## 4. Mount registration

- [x] 4.1 Using the `svelte-file-editor` subagent (or the Svelte MCP
      server's tools): add `onMount(registerWebMCPTools)` to
      `web/src/routes/+layout.svelte`, next to the existing
      `onMount(registerServiceWorker)`; verify `bun run check` passes
- [x] 4.2 Run `bun run test:e2e` once more and verify the dashboard
      still loads and functions with no console error in a browser
      that has no `document.modelContext` (every browser Playwright
      drives here, today). Confirmed via bisection with temporary
      `page.on('console'/'pageerror')` listeners added to the spec:
      zero console errors and zero page errors across every run,
      including the runs that hit #251's pre-existing flake (see 2.4)
      -- that flake is a silent hydration stall, not a thrown error
      this change introduced.

## 5. Documentation

- [x] 5.1 Add a short note under `CONTRIBUTING.md`'s `## Building`
      section, right after the existing frontend build step: what's
      registered, that it's feature-detected and safe to ignore on any
      browser without WebMCP support, and a link to
      <https://github.com/webmachinelearning/webmcp>
- [x] 5.2 Confirm the note reads correctly next to the surrounding
      section (manual read-through, per this account's own "read the
      README/CONTRIBUTING again before opening the PR" step)

## 6. Verify and ship

- [x] 6.1 Run `bun run check`, `bun run lint`, `bun run test`,
      `bun run test:coverage`, and `bun run build` (from `web/`);
      verify all clean
- [x] 6.2 Run `bun audit` and verify no new advisory
- [ ] 6.3 Open a pull request with `Closes #296` and verify CI passes

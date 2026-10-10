# Design

## Context

See proposal.md for motivation. Relevant current state:

- Every data-fetching page (`web/src/routes/+page.svelte`,
  `web/src/routes/pipelines/[id]/+page.svelte`,
  `web/src/routes/repos/[id]/usage/+page.svelte`) calls `fetch('/api/...')`
  inline inside a component function that also owns the page's local
  `$state` (the fetched data, an `error` string, pagination offsets) —
  there's no shared fetch module today to point a tool at.
- `web/src/routes/+layout.svelte` already runs one browser-feature-gated,
  fire-and-forget registration on mount:
  `onMount(registerServiceWorker)`, where `registerServiceWorker`
  (`web/src/lib/serviceWorker.ts`) checks `'serviceWorker' in navigator`
  before touching it and swallows failure ("best effort -- an install
  prompt just won't appear"). This is the shape to match.
- WebMCP is `document.modelContext.registerTool(tool, options)`, per the
  canonical spec repo and the W3C Community Group's own technical notes
  (both checked directly, in agreement with each other; several
  third-party blog posts found during research incorrectly describe a
  `navigator.modelContext` global, which isn't what either primary
  source defines):
  - <https://github.com/webmachinelearning/webmcp>
  - <https://w3c-cg.github.io/aikr/webMCP/webmcp-technical-notes.html>
  - <https://webmachinelearning.github.io/webmcp/>

  `tool` is `{name, description, title?, inputSchema?, execute,
  annotations?}`; `execute(input, {signal}) => Promise<{content: [...]}>`;
  `options.signal` is an `AbortSignal` that unregisters the tool when
  aborted. This is a pre-stable, actively-changing draft (experimental in
  Chrome only as of the sources above, no browser has shipped it stably)
  -- treat the exact shape as worth re-checking against the live spec at
  implementation time rather than trusted from this document alone.
- TypeScript's bundled DOM types don't know about `document.modelContext`
  yet (it isn't in any released `lib.dom.d.ts`), so `strict: true` in
  `web/tsconfig.json` will fail `bun run check` on `document.modelContext`
  without an ambient declaration.

## Goals / Non-Goals

**Goals:**

- Register `list_pipelines`, `get_pipeline`, and `get_repo_usage` as
  WebMCP tools, each backed by the same function the corresponding page
  already uses to fetch its data.
- Never break or degrade the dashboard in a browser without WebMCP
  support -- registration has to be purely additive.

**Non-Goals:**

- No write/mutating tool (nothing on these views mutates anything
  regardless).
- No attempt to polyfill or shim WebMCP for browsers that don't have it.
  A no-op there is the whole story.
- No change to what a page renders or how it's styled -- this only adds
  a second consumer of data the page already fetches.

## Decisions

### Extract one fetch function per endpoint into `web/src/lib/dashboardApi.ts`

Each becomes a small, page-independent `async function fetchX(...):
Promise<XResponse>` that does the `fetch('/api/...')` call, checks
`res.ok`, and returns the parsed body (or throws on failure) -- no
`$state` writes, no UI concerns. The three pages call these instead of
fetching inline; each page keeps owning its own `$state`, `error`
string, and how it reacts to a thrown error (existing behavior,
unchanged). The WebMCP tools call the exact same functions.

**Alternative considered**: give the WebMCP tool's `execute()` its own
`fetch('/api/...')` call, leaving the pages untouched. Rejected --
that's precisely the "second path to the same API" the issue's own
approach section warns against, and it means the tool's shape silently
drifts from the page's the first time either one's request changes
(a new query parameter, a renamed field) without the other.

### Tool registration in `web/src/lib/webmcpTools.ts`, called from `+layout.svelte`

`registerWebMCPTools(): void` mirrors `registerServiceWorker`'s shape:
a feature check (`if (!('modelContext' in document)) return;`) then
three `document.modelContext.registerTool(...)` calls, each `execute`
calling one `dashboardApi.ts` function and shaping its result as
`{content: [{type: 'text', text: JSON.stringify(data)}]}`. Wired with
`onMount(registerWebMCPTools)` in `+layout.svelte`, next to the existing
`onMount(registerServiceWorker)` -- registered once per page load, same
as the service worker, not re-registered on every client-side
navigation.

**Alternative considered**: register per-page, inside each of the three
`+page.svelte` files, so a tool only exists while its page is mounted.
Rejected for v1 -- simpler to reason about with everything registered
once at the layout level, and nothing in the issue's acceptance
criteria asks for page-scoped tool availability. Worth revisiting if a
future tool count makes "every tool, always available" noisy for an
agent (open question below).

### Ambient type declaration for `document.modelContext`

A `ModelContextTool`/`Document.modelContext` ambient interface, scoped
to `web/src/lib/webmcpTools.ts`'s own needs (not the full spec surface),
declared via `declare global { interface Document { modelContext?:
... } }` in a new `web/src/lib/webmcp.d.ts` -- kept separate from
`web/src/app.d.ts`, which is SvelteKit's own generated-shape file, not
the place for a third-party/experimental browser API's types.

### `.svelte` file changes go through the svelte-file-editor agent

This account's session guidance requires the `svelte-file-editor`
subagent (or the Svelte MCP server's own tools) for creating, editing,
or reviewing any `.svelte` file. `+layout.svelte`'s one-line addition
(`onMount(registerWebMCPTools)` plus an import) and the three pages'
switch from inline `fetch` to calling `dashboardApi.ts` are the only
`.svelte` edits this change makes -- an implementation-phase
constraint on *how* those edits get made, not a design decision about
what changes.

### Where the CONTRIBUTING.md note goes

Under `## Building`, right after the existing frontend build step,
rather than under "## Testing and linting"'s `Frontend (from web/):`
lead-in -- the issue's "documented alongside the existing frontend
build instructions" reads more naturally next to *what the built app
does* than next to how it's tested. One or two sentences: what's
registered, that it's feature-detected and safe to ignore, and a link
to the WebMCP spec repo for anyone unfamiliar with it.

## Risks / Trade-offs

- **API still in flux** -- a future draft could rename
  `registerTool`, change `execute`'s signature, or move off
  `document.modelContext` entirely. Mitigated by isolating every
  WebMCP-specific call in `webmcpTools.ts` and its own `.d.ts` -- a spec
  change is a one-file update, not a hunt through page components.
- **No automated test can exercise a real WebMCP agent call** -- no
  browser in this project's Playwright/e2e setup implements
  `document.modelContext` yet. Mitigated by testing what's actually
  ours to guarantee: `registerWebMCPTools` calls `registerTool` with the
  right arguments when `document.modelContext` exists (a fake object
  standing in for the browser API), and does nothing when it doesn't --
  a unit test (`bun:test`), not an e2e one. The real end-to-end
  guarantee — an agent inside an actual WebMCP browser gets a correct
  answer — isn't testable here until a browser or test harness
  implements the API.
- **Registered tools are global to the app, not scoped to what's on
  screen** -- calling `get_pipeline` while the user is looking at the
  overview page still works (the underlying fetch doesn't care), which
  is a feature for an agent, not a risk, but worth naming since it's a
  deliberate choice (see the per-page-registration alternative above).

## Open Questions

- Whether tool registration should become page-scoped once the app has
  enough views that "every tool, all the time" gets noisy for an agent
  browsing tool lists. Doesn't change this change's specs, approach, or
  tasks -- a future change's call once there's a concrete signal it's a
  problem.

# Proposal

## Why

After #523's cache and compression fixes, the dashboard's 80 KB global
stylesheet still blocked first render and the Inter font was fetched only once
that sheet had loaded. `render-blocking-insight` and
`network-dependency-tree-insight` scored 0 on every page. Tracked in #527.

## What Changes

- SvelteKit inlines the stylesheet into the SPA shell (`inlineStyleThreshold`).
- The latin Inter subset is preloaded from a `handle` hook, so it starts with
  the document, not after the CSS.
- The web manifest link is added after `load`, so it isn't a critical request.
- Both insights are asserted at `error` instead of `warn`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci-pipeline`: adds a requirement that Lighthouse gates render-blocking
  requests and request chains.

## Impact

- `web/vite.config.ts`, `web/src/hooks.server.ts`, `web/src/app.html`,
  `web/lighthouserc.cjs`.

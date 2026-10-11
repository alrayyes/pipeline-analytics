## Context

`web/` is a client-routed SPA built with adapter-static and `fallback:
'index.html'`, so the server can't tell a route from a typo without a list.
Dynamic routes (`/runs/[id]`) rule out prerendering every path.

## Decisions

- **The build writes the route list.** `web/scripts/generate-routes.ts` walks
  `src/routes` for directories holding a `+page.*` and writes `routes.json`
  (patterns such as `/runs/[id]`) into the embedded `dist/`. Reading the same
  tree the router reads means a new route can't be forgotten, which a
  hand-kept Go list would allow. Chosen over asking SvelteKit for its manifest:
  adapter-static exposes it only inside the adapter, and a postbuild script is
  a few lines.
- **Matching.** `[name]` matches one non-empty segment; everything else is
  literal; `path.Clean` first. Layout-only and `+server` directories aren't
  routes. `/steps` stays a route: its client-side redirect needs the shell.
- **Missing `routes.json` means "serve the shell".** A checkout without a
  frontend build has no `index.html` either; the Go tests use their own
  in-memory assets and shouldn't need `bun run build`.
- **A `404` carries the app shell, not the static page.** The not-found view
  belongs in the shared layout (navigation, theme), which only the app can
  render, and the router already shows it for a path no route matches. The
  status is the server's to fix; the body stays the shell. `404.html` remains
  for when the shell can't be read. A missing `_app/` file gets a plain `404`,
  since HTML there only fails to run.
- **`500` is for the server's own failure** (an unreadable asset), not for the
  app's runtime errors, which `+error.svelte` renders (#576). If `500.html`
  itself is unreadable the body is plain text with the status.

## Risks

- A route added with a hand-written server `+page` outside `src/routes` would
  be missed; the repo has none, and the route-list test covers the tree.

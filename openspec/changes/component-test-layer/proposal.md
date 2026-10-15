# Proposal

## Why

`web/` has no layer between a unit test and a browser, so a page's own wiring
(what its window toggle asks the server for, what it says when there is
nothing or the request fails) could only be tested in Playwright. Four pages
repeat those tests, each paying for a real server and a sign-in (#614).
`rules/testing.md` says to assert a state where its logic lives and to keep one
journey per flow.

## Research

- The Svelte docs say that testing components client-side needs a DOM
  environment, show jsdom with `@testing-library/svelte`, and recommend Vitest
  for a Vite project ([Svelte docs, Testing](https://svelte.dev/docs/svelte/testing)).
  They do not mention happy-dom or Bun.
- This repo already runs `bun test`, with a plugin that compiles `.svelte.ts`
  modules. Vitest would be a second test runner, which `testing.md` rules out, so
  the layer is built on `bun test` with Svelte's own `mount`.
- The DOM is happy-dom, registered as globals by its own
  `@happy-dom/global-registrator` package, which describes itself as for a Node
  environment. It works under Bun, which I checked by running it: the Failures
  page mounts and its toggle, request and states behave. I chose it over jsdom
  for being lighter; I did not benchmark the two.
- Svelte must resolve to its client entry, which needs Bun's `browser` export
  condition (`--conditions browser`); without it `mount` throws "not available
  on the server". Found by running it.
- Rejected: Vitest or Playwright component testing, each a second runner or a
  browser per run.

## What Changes

- A Bun plugin compiles `.svelte` components, beside the `.svelte.ts` modules it
  already compiles.
- `test/dom.ts` registers a happy-dom window; a test file opts in by importing
  it first, so the plain unit tests still run without one.
- `bun run test`, `test:coverage` and Stryker run with `--conditions browser`.
  The root layout test imports the DOM helper, since SvelteKit's `redirect()`
  reads `window` under that condition.
- The Failures page gets four component tests, and the three e2e tests they
  replace are deleted (window toggle, nothing and error, no saved window).

## Capabilities

### Modified Capabilities

- `ci-pipeline`: a requirement that page wiring is tested without a browser.

## Impact

- `web/test/`, `web/package.json`, `web/stryker.conf.mjs`,
  `web/src/routes/failures/`, `web/tests/e2e/failures.spec.ts`,
  `CONTRIBUTING.md`; `@happy-dom/global-registrator` as a dev dependency.

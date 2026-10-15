# Proposal

## Why

`/releases` renders every release: 114 cards and about 2,500 elements, which
Lighthouse's `dom-size` flags (#618). On one machine, back to back, main's page
scored 72 with 1,010 ms of blocking time and 4.6 s of main-thread work; the
published CI run scores it 93 with the same 2,505 elements.

## What Changes

- The page opens on the newest 20 releases, with a "Show older releases"
  button that adds 20 more until none are left.
- The first 20 still render slice by slice, so opening the page is not one long
  task. Asking for older releases renders those 20 at once, which is far fewer
  than the 114 that were rendered together before.
- The release journey in `dashboard.spec.ts` asserts the cap and pages through
  to reach an old release.

## Capabilities

### Modified Capabilities

- `dashboard-ui`: "The release history renders in slices" no longer shows every
  release at once.

## Impact

- `web/src/routes/releases/+page.svelte`, `web/tests/e2e/dashboard.spec.ts`.

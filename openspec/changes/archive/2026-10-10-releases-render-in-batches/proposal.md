# Proposal

## Why

`/releases` scored 0.45 on Lighthouse performance, the lowest page. The cause
is `total-blocking-time`: the page parsed and sanitised every release's
markdown in one task (a 660 ms long task for 98 releases), and the list grows
with each release. Tracked in #522.

## What Changes

- The release list renders five releases at once, then five more per task
  until all are shown. Every release is still on the page.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement that the release history renders without
  a long main-thread task.

## Impact

- `web/src/routes/releases/+page.svelte`, new `web/src/lib/batches.ts`.

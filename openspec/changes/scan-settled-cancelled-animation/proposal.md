# Proposal

## Why

`scanSettled` (#570) waits on `Animation.finished`, which rejects with an
AbortError when the animation is cancelled mid-wait, such as a dialog closing.
That failed the release commit's e2e (#573).

## What Changes

- A cancelled animation counts as settled: the wait ignores the rejection and
  the scan carries on.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none; test-only change, no product behaviour changes)

## Impact

- `web/tests/e2e/axe.ts` and a regression test, `scan-settled.spec.ts`.

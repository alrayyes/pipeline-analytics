# Proposal

## Why

Two pages moved content the visitor was looking at once their data arrived.
Lighthouse measured it on a phone-sized viewport (#584): 0.063 of layout shift
on `/releases`, where the footer sat under "Loading…" and was pushed away by
the list, and 0.029 on `/settings`, where the passkey rows pushed the cards
below them down.

## What Changes

- The release history's loading state is a screen tall, so the footer starts
  below the fold and the list pushes it from out of sight.
- The Passkeys table shows a one-row "Loading…" placeholder of the height of
  a real row until the first load lands.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement that late-arriving content doesn't shift what is on screen.

## Impact

- `web/src/routes/releases/+page.svelte`, `web/src/routes/settings/+page.svelte`
  and `web/tests/e2e/layout-shift.spec.ts`.

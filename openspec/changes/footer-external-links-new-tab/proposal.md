# Proposal

## Why

The footer's GitHub and license links replaced the dashboard in the current
tab, so reading the repository meant losing your place. Tracked in #554.

## What Changes

- The GitHub and license links open in a new tab, with
  `rel="noopener noreferrer"` and visually hidden "(opens in a new tab)" text.
- The version (`/releases`) and Privacy & disclaimer (`/legal`) links are the
  app's own pages and stay in the same tab.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement for which footer links leave the app.

## Impact

- `web/src/lib/components/Footer.svelte`, `web/tests/e2e/footer.spec.ts`.

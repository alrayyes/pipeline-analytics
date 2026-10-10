# Proposal

## Why

Lighthouse's accessibility audits flagged two things the axe scan doesn't:
`label-content-name-mismatch` on the time-window buttons (`/failures`,
`/flaky`: visible "7d", accessible name "7 days") and `heading-order` on
`/settings` (an `h3` straight under the `h1`).

## What Changes

- The time-window buttons show their full label ("24 hours", "7 days",
  "30 days"), so the visible text is the accessible name.
- The "Save a token" heading on Settings is an `h2`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement that visible control text is contained in
  the accessible name.

## Impact

- `web/src/lib/components/telemetry/WindowToggle.svelte`,
  `web/src/lib/components/SavedTokens.svelte`.

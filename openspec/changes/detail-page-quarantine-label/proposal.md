# Proposal

## Why

The flaky list labels a quarantined step (#541), but the pipeline detail
page's steps table only says "flaky", so a step someone has already handled
looks unhandled there. Tracked in #551.

## What Changes

- The steps table on the pipeline detail page shows a "Quarantined" label, in
  words, beside the "flaky" marker of a quarantined step.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement for the detail page's label.

## Impact

- `web/src/routes/pipelines/[id]/+page.svelte`, `web/src/lib/dashboardApi.ts`
  (`Step.quarantined`), `web/tests/e2e/detail-pages.spec.ts`.

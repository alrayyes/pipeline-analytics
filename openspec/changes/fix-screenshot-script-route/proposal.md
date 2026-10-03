# Proposal

## Why

The Pipelines list moved from `/` to `/pipelines` and `/` became the failure
overview. The capture script still loaded `/` and waited for the Pipelines
page's health filter, so the release workflow's screenshot job timed out on
the 0.68.0 and 0.69.0 releases. Tracked in #411.

## What Changes

- The capture script drives `/pipelines` for the Pipelines overview, light
  and dark.
- The `docs-screenshots` spec says the script follows the route the page
  lives at, and that a full run succeeds.

## Capabilities

### Modified Capabilities

- `docs-screenshots`: the script's first step names the Pipelines route.

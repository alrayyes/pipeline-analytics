# Proposal

## Why

A broken feature branch hides a healthy main in the telemetry views. The run
branch has been recorded since #335 but nothing filters on it. Tracked in
#344, backend half. The frontend session builds the selector against this.

## What Changes

- `GET /api/insights/failures`, `GET /api/runs` and `GET /api/steps/flaky`
  take an optional `branch`, matched exactly, at most 255 characters.
  Omitted covers every branch, as before.
- The MCP tools for those three take the same `branch`.
- A branch with no runs gives an empty result, not an error.
- Listing the branches a selector offers is a separate change.

## Capabilities

### Modified Capabilities

- `pipeline-metrics`: telemetry can be scoped to one branch.

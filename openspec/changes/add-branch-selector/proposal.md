# Proposal

## Why

A broken feature branch hides a healthy main in the telemetry views. The
backend half of #344 (`branch-filter`) lets the three reads take a `branch`;
nothing in the dashboard sends it yet.

## What Changes

- A branch selector beside the time window on the overview, the root-cause,
  the flaky and the runs views. It offers the branches that have runs in the
  window, busiest first, from `GET /api/branches`.
- The choice is view state, shared by those four views and kept for the visit.
  It is not an account setting, so nothing is added to `/api/settings`.
- No selection means every branch, as today. A branch with no runs shows an
  empty state that names the branch, not zeros.

## Capabilities

### Modified Capabilities

- `dashboard-ui`: telemetry views can be scoped to one branch.

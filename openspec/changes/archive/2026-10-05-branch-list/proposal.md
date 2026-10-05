# Proposal

## Why

The branch filter (#458) needs something to offer: a selector has to know
which branches have runs. Tracked in #344, backend half. Stacked on #458.

## What Changes

- `GET /api/branches` lists the branches with runs in a 24h, 7d or 30d
  window, each with its run count, busiest first then by name. A run with no
  recorded branch isn't a branch. Same repo, forge and window parameters as
  the insights.
- An MCP tool, `list_branches`, returns the same.

## Capabilities

### Modified Capabilities

- `pipeline-metrics`: the branches to scope telemetry by can be listed.

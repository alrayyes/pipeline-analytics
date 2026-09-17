# Proposal

## Why

The Release workflow's "refresh README screenshots" job has failed on
the last two runs because `capture-screenshots.mjs` still targets a
button PR #158 removed. Tracked as
[alrayyes/pipeline-analytics#182](https://github.com/alrayyes/pipeline-analytics/issues/182).

## What Changes

- Update `web/scripts/capture-screenshots.mjs` to click the health
  filter's "All" radio, scoped to its `"Filter by health status"`
  radiogroup (an unscoped selector also matches `ForgeFilter`'s own
  "All" radio).

## Capabilities

None of the four existing capabilities cover this — it's a docs/CI
tooling script, not dashboard behavior. New, narrow capability for
keeping that script in sync with the UI it drives.

### New Capabilities

- `docs-screenshots`: the script that captures the README's
  screenshots against a real running server stays able to drive the
  current UI.

### Modified Capabilities

(none)

## Impact

- `web/scripts/capture-screenshots.mjs` (two selector fixes).
- No application code, no user-facing behavior.

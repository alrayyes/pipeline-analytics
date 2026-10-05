# Proposal

## Why

The frontend mutation job reported a score and never failed, so a test that
doesn't pin behaviour could merge. The score was 67%. Tracked in #422.

## What Changes

- `stryker.conf.mjs` sets `thresholds.break` to 100, as `rules/javascript.md`
  asks.
- Tests kill the surviving mutants in the forge filter and Pipelines filters
  stores: cache keys, the cached read, the server patch bodies, and the
  initial values.
- Two mutants are equivalent and carry a Stryker disable comment saying why.
- `bun test` runs with `--isolate`, in the scripts and for Stryker, so a test
  that reads a store's initial value doesn't depend on which file ran first.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: a surviving frontend mutant fails the build.

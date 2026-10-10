# Proposal

## Why

`lefthook.yml` kept its own globs next to the CI filter in `changes.sh`, so
the two disagreed: the Go hooks skipped `go.sum` and `openapi/**`, and the
frontend hooks skipped `CHANGELOG.md`, all of which CI runs on. A push could
pass locally and fail in the pipeline. Tracked in #374; follows #371.

## What Changes

- Each `pre-push` job, and the pre-commit `docker build`, runs through
  `.github/scripts/hook-guard.sh`, which asks `changes.sh` whether the
  job's CI group changed.
- The hook globs for those jobs are gone, so `changes.sh` is the only list
  of paths.
- An unknown group name fails the hook rather than skipping it.
- Pre-commit fixers keep a file-type glob: they act on the staged files.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: local hooks follow the same groups as CI.

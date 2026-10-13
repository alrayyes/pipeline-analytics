# Proposal

## Why

`deploy API docs` failed twice on 2026-10-13 at `actions/deploy-pages` with
"No artifacts named github-pages were found" (#606). The upload step had
finished 0.3 seconds earlier. The run's artifact list lags an upload, and
`deploy-pages` reads it once. A full re-run passed each time; re-running only
the failed job fails again, because it uploads a second `github-pages`.

## What Changes

- `wait-for-artifact.sh` polls the run's artifact list until it names
  `github-pages`, for up to a minute, and fails with a message if it never does.
- The deploy job runs it between the upload and `deploy-pages`, and gets
  `actions: read` to list the artifacts.
- `wait-for-artifact.test.sh` covers it, and the CI script-test step runs that.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci-pipeline`: adds a requirement that the Pages deploy waits for its
  artifact.

## Impact

- `.github/workflows/ci.yml`, `.github/scripts/wait-for-artifact.sh` and its
  test.

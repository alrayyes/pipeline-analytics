# Proposal

## Why

Codecov answers an upload with a 403 now and then: 4 of the last 30 CI runs on
main failed that way, including the 0.78.1 release commit (#602). Every
re-run passed. A failed upload says nothing about the code, but all four
Codecov steps set `fail_ci_if_error: true`, so it turns main red.

## What Changes

- The four Codecov steps in `ci.yml` set `fail_ci_if_error: false`. An upload
  error becomes a warning in the log. The test and coverage steps before them
  still fail the job when the code is wrong.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci-pipeline`: adds a requirement that a Codecov upload error doesn't fail
  the build.

## Impact

- `.github/workflows/ci.yml`.

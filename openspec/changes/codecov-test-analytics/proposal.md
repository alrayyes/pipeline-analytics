# Proposal

## Why

CI uploads coverage to Codecov but not test results, so Codecov can't show
flaky tests or failure history. Tracked in #421 (criteria 1 and 2; bundle
analysis is its own change).

## What Changes

- The Go job runs the tests through `gotestsum --junitfile`.
- The frontend's `test:coverage` script also writes JUnit.
- Each job uploads its JUnit file with `report_type: test_results`, after a
  failing test too, and skips the upload on Dependabot runs as the coverage
  upload does.
- A test in the `changes` job fails on a hyphenated `report-type`, a missing
  `!cancelled()` or a suite that stops writing JUnit.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: test results reach Codecov for both suites.

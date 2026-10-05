# Spec Delta

## ADDED Requirements

### Requirement: Telemetry can be scoped to one branch

The system SHALL accept an optional `branch` on the failure insights, the run
list and the flaky steps, cover only runs on that exact branch when it is
given and every branch when it is not, and reject one longer than 255
characters.

#### Scenario: A healthy main beside a broken feature branch

- **WHEN** the insights are read with `branch=main`
- **THEN** only main's runs count toward the totals and the pass rate

#### Scenario: A branch with no runs

- **WHEN** a branch has no runs in the window
- **THEN** the answer is an empty result, not an error

#### Scenario: A branch that is too long

- **WHEN** `branch` is longer than 255 characters
- **THEN** the answer is `400`

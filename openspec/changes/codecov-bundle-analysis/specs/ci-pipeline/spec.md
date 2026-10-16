# Spec Delta

## ADDED Requirements

### Requirement: A frontend build uploads bundle analysis to Codecov

The system SHALL upload bundle analysis to Codecov from the frontend build in
CI, using the same `CODECOV_TOKEN` as the coverage and test result uploads. A
build with no token SHALL still succeed and upload nothing.

#### Scenario: A build in CI

- **WHEN** the `frontend` job builds the web app with `CODECOV_TOKEN` set
- **THEN** the build uploads its bundle analysis

#### Scenario: A build with no token

- **WHEN** the web app is built locally, on a fork or on a Dependabot run
- **THEN** the build succeeds and nothing is uploaded

#### Scenario: The plugin is removed

- **WHEN** `vite.config.ts` no longer registers the Codecov plugin
- **THEN** the `changes` job fails

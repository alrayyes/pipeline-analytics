# Spec Delta

## ADDED Requirements

### Requirement: Local hooks run on the files CI runs on

The system SHALL decide whether a git hook runs from the same file groups as
the CI job that runs the same tool, and SHALL fail a hook that names a group
the filter doesn't define.

#### Scenario: A spec-only change runs the Go tests locally

- **WHEN** a push changes only `openapi/openapi.yaml`
- **THEN** the Go test hook runs, as the CI job does

#### Scenario: A docs-only change skips code hooks

- **WHEN** a push changes only documentation
- **THEN** no Go or frontend hook runs

#### Scenario: A hook names a group that doesn't exist

- **WHEN** a hook asks for a group `changes.sh` doesn't define
- **THEN** the hook fails instead of skipping

# Spec Delta

## ADDED Requirements

### Requirement: CI tools run on the bun version CI pins

The system SHALL keep each frontend dev tool at a release that runs on the bun
version the CI jobs pin, and SHALL stop Dependabot from proposing a release
that needs a newer bun.

#### Scenario: A web change runs the mutation job

- **WHEN** a pull request changes a file under `web/`
- **THEN** the mutation job starts Stryker and ends on its mutation score, not
  on a bun version check

#### Scenario: A tool release needs a newer bun

- **WHEN** a release of `@hughescr/stryker-bun-runner` needs bun 1.4 or newer
- **THEN** Dependabot does not propose it while CI pins bun 1.3

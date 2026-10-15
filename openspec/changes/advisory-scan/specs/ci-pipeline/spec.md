# Spec Delta

## MODIFIED Requirements

### Requirement: The dependency audit has no ignores

The system SHALL run the dependency audits, `bun audit` and `govulncheck`, in CI
without any ignore list, and SHALL fail a pull request only on an advisory that
it introduces over its base. A scan of the default branch SHALL keep one open
issue per lockfile that lists the advisories found, so that a known advisory
with no fix stays visible instead of being hidden.

#### Scenario: A pull request that adds an advisory

- **WHEN** a pull request changes the manifest or lockfile and the audit then
  reports an advisory the base did not have
- **THEN** the audit job fails

#### Scenario: An advisory published overnight

- **WHEN** a new advisory is published against a dependency a pull request does
  not touch
- **THEN** the pull request's audit job passes

#### Scenario: An advisory with no fix

- **WHEN** a dependency on the default branch has an advisory and no patched
  release
- **THEN** the scan lists it in its open `Dependency advisories` issue, and no
  ignore is added to the pipeline

#### Scenario: A clean audit

- **WHEN** the scan runs and the audit reports nothing
- **THEN** the issue it kept is closed

# Spec Delta

New capability `ci-pipeline`: decides which CI checks run for a given
change, so the pipeline spends runner time only on checks that can fail.

## ADDED Requirements

### Requirement: Checks run only when the files they cover change

The system SHALL run each CI check only when a file it covers, the tool's own
configuration or lockfile, or the pipeline definition itself changed, and
SHALL skip it otherwise without failing the pull request.

#### Scenario: A docs-only change skips code checks

- **WHEN** a pull request changes only documentation
- **THEN** the Go, frontend, end-to-end, Lighthouse, Docker and spec checks
  are skipped and the prose and markdown checks run

#### Scenario: A code-only change skips prose checks

- **WHEN** a pull request changes only Go source
- **THEN** the prose and markdown checks are skipped and the Go checks, the
  end-to-end check and the Docker build run

#### Scenario: A skipped required check does not block the merge

- **WHEN** a required check is skipped because nothing it covers changed
- **THEN** the pull request is still mergeable

#### Scenario: Editing the pipeline runs every check

- **WHEN** a pull request changes the CI workflow or its filter script
- **THEN** every check runs

#### Scenario: An unknown base runs every check

- **WHEN** the changed files can't be determined, as on a new branch
- **THEN** every check runs

### Requirement: Some workflows are never filtered

The system SHALL run the release, pull-request title lint and merge
automation workflows regardless of which files changed, and SHALL keep at
least one always-running job so every pull request has a status.

#### Scenario: A docs-only push to main still releases

- **WHEN** a push to main changes only documentation
- **THEN** the release workflow still runs

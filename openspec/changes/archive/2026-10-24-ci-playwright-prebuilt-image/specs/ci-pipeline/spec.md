# Spec Delta

## ADDED Requirements

### Requirement: The CI Playwright image matches the lockfile

The system SHALL run the Playwright jobs in the `ci-playwright` image pinned by
exact version and digest, SHALL keep that version equal to the
`@playwright/test` version in `web/bun.lock`, and SHALL fail a check when they
differ.

#### Scenario: The image is older than the lockfile

- **WHEN** `web/bun.lock` locks `@playwright/test` 1.64.1 and a workflow uses
  the `ci-playwright:1.63.0` image
- **THEN** the `changes` job and the pre-push hook fail, naming both versions

#### Scenario: The image is not pinned by digest

- **WHEN** a workflow names the image by tag alone, or two workflows use
  different digests
- **THEN** the check fails

#### Scenario: A job installs the browsers itself

- **WHEN** a workflow runs `playwright install`, or the container lacks
  `--ipc=host`
- **THEN** the check fails

# Spec Delta

## ADDED Requirements

### Requirement: The API docs deployment waits for its artifact

The system SHALL wait until the run's artifact list names the `github-pages`
artifact before it deploys. A list that trails the upload then doesn't fail the
deployment.

#### Scenario: The list trails the upload

- **WHEN** the deployment job starts while the artifact list omits
  `github-pages`
- **THEN** the job looks again for up to a minute, then deploys once it shows

#### Scenario: The artifact never shows

- **WHEN** the list still doesn't name `github-pages` after a minute
- **THEN** the job fails and its log says the artifact is missing

# Spec Delta

## ADDED Requirements

### Requirement: The Pages deploy waits for its artifact

The system SHALL wait until the run's artifact list names the Pages artifact
before it deploys, so that a list that trails the upload doesn't fail the
deploy.

#### Scenario: The list trails the upload

- **WHEN** the deploy job starts and the artifact list doesn't name
  `github-pages` yet
- **THEN** the job looks again for up to a minute, then deploys once it shows

#### Scenario: The artifact never shows

- **WHEN** the list still doesn't name `github-pages` after a minute
- **THEN** the job fails and its log says the artifact is missing

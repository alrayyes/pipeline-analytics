# Spec Delta

## ADDED Requirements

### Requirement: A tracked repository can be fetched by id

The system SHALL return one tracked repository at `GET /api/repos/{repoId}`,
with its token masked as the list masks it, and SHALL answer 404 with the API's
error body when no repository is tracked under the id.

#### Scenario: A tracked id

- **WHEN** a signed-in caller requests a tracked repository's id
- **THEN** the response is that repository with its masked token and never the
  token itself

#### Scenario: An unknown id

- **WHEN** a caller requests an id no repository is tracked under
- **THEN** the response is 404 with the `not_found` error body

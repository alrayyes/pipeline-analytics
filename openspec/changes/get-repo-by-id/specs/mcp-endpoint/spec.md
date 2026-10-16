# Spec Delta

## ADDED Requirements

### Requirement: A tool gets one tracked repository

The system SHALL expose a `get_repo` MCP tool that returns one tracked
repository by id, calling the same code as `GET /api/repos/{repoId}`.

#### Scenario: A tracked id

- **WHEN** an agent calls `get_repo` with a tracked repository's id
- **THEN** it gets that repository with its masked token only

#### Scenario: An unknown id

- **WHEN** an agent calls `get_repo` with an id no repository is tracked under
- **THEN** the tool reports a not-found error, as the endpoint answers 404

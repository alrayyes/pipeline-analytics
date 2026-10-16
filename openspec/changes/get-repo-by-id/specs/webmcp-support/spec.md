# Spec Delta

## ADDED Requirements

### Requirement: A tool gets one tracked repository

The system SHALL expose a `get_repo` WebMCP tool that returns one tracked
repository by id, sending the request `GET /api/repos/{repoId}`.

#### Scenario: A tracked id

- **WHEN** an in-browser agent calls `get_repo` with a tracked repository's id
- **THEN** it gets that repository, requested by its id

#### Scenario: An unknown id

- **WHEN** the agent calls `get_repo` with an id no repository is tracked under
- **THEN** the tool fails saying the repository was not found

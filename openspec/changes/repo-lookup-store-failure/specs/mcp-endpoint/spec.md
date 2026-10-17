# Spec Delta

## ADDED Requirements

### Requirement: get_repo reports a store failure as an error

The `get_repo` MCP tool SHALL report a not-found error only for an id no
repository is tracked under. A store failure SHALL be reported as an error that
isn't a not-found, as the endpoint answers 500.

#### Scenario: A store failure

- **WHEN** an agent calls `get_repo` and the store fails
- **THEN** the tool reports an internal error, not a not-found

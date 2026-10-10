# Spec Delta

## ADDED Requirements

### Requirement: Every read endpoint has a tool

The system SHALL expose each read endpoint that serves pipeline data as an
MCP tool returning the same data, and SHALL fail its tests when a GET
operation in the OpenAPI description has neither a tool nor a stated reason
for having none.

#### Scenario: A run an agent finds can be opened

- **WHEN** an agent has a run id from `list_runs` or
  `list_pipeline_flaky_runs`
- **THEN** `get_run_steps` returns that run's steps with their outcomes

#### Scenario: A new endpoint without a tool

- **WHEN** a GET operation is added to the OpenAPI description with no tool
  and no exclusion entry
- **THEN** the test suite fails and names the operation

#### Scenario: Failure insights over MCP

- **WHEN** an agent calls `get_failure_insights` with no window
- **THEN** the result reports the window the server used

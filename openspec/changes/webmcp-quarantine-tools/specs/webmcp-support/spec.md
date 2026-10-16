# Spec Delta

## ADDED Requirements

### Requirement: Tools mark and clear a step's quarantine

The system SHALL expose as WebMCP tools marking a flaky step as quarantined,
with an optional note, and clearing that mark, each sending the same request as
the Flaky page's control for it.

#### Scenario: Marking a step

- **WHEN** an in-browser agent calls `quarantine_step` with a pipeline, a step
  and a note
- **THEN** the same PUT as the Flaky page's mark control is sent, with the step
  escaped, and the result is the quarantine state the server returns

#### Scenario: A missing argument

- **WHEN** an agent calls `quarantine_step` or `unquarantine_step` without a
  `pipelineId` or a `step`
- **THEN** the call fails saying which is required and nothing is sent

#### Scenario: The server refuses

- **WHEN** the server answers the request with an error, such as 422 for a
  note that is too long
- **THEN** the call fails and carries the status

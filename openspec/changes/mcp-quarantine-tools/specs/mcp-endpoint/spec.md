# Spec Delta

## ADDED Requirements

### Requirement: Tools can quarantine and un-quarantine a flaky step

The system SHALL expose as MCP tools marking a step as known for 30 days, with
an optional note of at most 500 characters, and clearing that mark, each acting
only for a signed-in session and reporting a failure as the endpoint does.

#### Scenario: Marking a step

- **WHEN** a client with a session calls `quarantine_step` for a flaky step
- **THEN** the step is quarantined for 30 days and the flaky list reports it

#### Scenario: A note that is too long

- **WHEN** a client with a session calls `quarantine_step` with a note over 500
  characters
- **THEN** the call reports `note_too_long` and nothing is marked

#### Scenario: An API token

- **WHEN** a client authenticated only by an API token calls `quarantine_step`
  or `unquarantine_step`
- **THEN** the call reports `unauthenticated` and the mark is unchanged

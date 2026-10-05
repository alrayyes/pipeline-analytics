# Spec Delta

## ADDED Requirements

### Requirement: A run can be re-run or cancelled on its forge

The system SHALL let a signed-in session re-run a concluded run, re-running only
the failed jobs when the run failed, and cancel a queued or running run, by
asking the run's forge, and SHALL answer `202` when the forge accepts it.

#### Scenario: Re-running a failed run

- **WHEN** a session re-runs a failed GitHub run
- **THEN** the failed jobs are re-run and the answer is `202`

#### Scenario: Cancelling a running run

- **WHEN** a session cancels a running GitHub run
- **THEN** the forge is asked to cancel it and the answer is `202`

### Requirement: A refused or impossible action says why

The system SHALL answer `404` for an unknown run, `409 not_actionable` when the
run's state doesn't allow the action, `403 forbidden` when the stored token
lacks write access to Actions, `501 unsupported` when the forge has no API for
it and `502 unreachable` when the forge doesn't answer.

#### Scenario: A read-only token

- **WHEN** the stored token only has read access to Actions
- **THEN** the answer is `403` with code `forbidden`, and the dashboard says the
  token needs Actions write permission

#### Scenario: A Forgejo run

- **WHEN** the run is on a Forgejo instance
- **THEN** the answer is `501` with code `unsupported`

### Requirement: Only a session may act on a run

The system SHALL answer `401` to an API token for these actions and SHALL NOT
offer them as MCP tools.

#### Scenario: An API token

- **WHEN** an API token asks to cancel a run
- **THEN** the answer is `401`

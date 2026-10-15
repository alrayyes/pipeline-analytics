# Spec Delta

## REMOVED Requirements

### Requirement: No write tool

**Reason**: Agents need the writes the API offers (`rules/mcp.md`). The rule
that mattered, that an agent holding only an API token must not change stored
state or act on a forge, is kept and stated below.

**Migration**: None. The read tools are unchanged.

## ADDED Requirements

### Requirement: Write tools act only for a signed-in session

The system SHALL let an MCP tool that writes, to the forge or to stored state,
act only when the call was authenticated by a session, and SHALL refuse a call
authenticated by an API token before the write is attempted, with the error the
matching endpoint gives.

#### Scenario: Re-running a failed run with a session

- **WHEN** a client with a session calls `rerun_run` for a run that failed
- **THEN** the forge is asked to re-run the failed jobs and the result reports
  the request as accepted

#### Scenario: An API token

- **WHEN** a client authenticated only by an API token calls `rerun_run` or
  `cancel_run`
- **THEN** the call reports `unauthenticated` and the forge is not asked

#### Scenario: A run that cannot be acted on

- **WHEN** a client with a session calls `rerun_run` for a run that has not
  concluded, or `cancel_run` for one that has
- **THEN** the call reports `not_actionable`, as the endpoint answers 409, and
  the forge is not asked

# Spec Delta

## ADDED Requirements

### Requirement: A job's log tail is read from the forge on demand

The system SHALL return the last lines of a job's log, 200 by default and at
most 1000, fetched from the forge when asked, never stored, with ANSI sequences
untouched.

#### Scenario: A GitHub job's log

- **WHEN** a signed-in user asks for a GitHub job's log
- **THEN** the server follows GitHub's redirect and returns the last lines
  oldest first, `truncated` true if there were more

### Requirement: A missing log says why and keeps the deep link

The system SHALL answer `200` with `available: false`, a `reason` and the
job's `forgeUrl` when the forge can't supply the log, and `404` for an unknown
run or job.

#### Scenario: A Forgejo job

- **WHEN** the job is on a Forgejo instance without a log API
- **THEN** `available` is false, `reason` is `unsupported` and `forgeUrl` is set

#### Scenario: An unknown job

- **WHEN** the run or job doesn't exist
- **THEN** the answer is `404`

# job-logs Specification

## Purpose
Lets the dashboard and agents read the end of a job's log through the API without the service storing logs, and says why when a forge can't supply one.

## Requirements

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

### Requirement: A step names its job

The system SHALL include the ID of the job a step ran in on every step it
returns. A client can ask for that job's log.

#### Scenario: A run's steps

- **WHEN** a client reads a run's steps
- **THEN** each step carries a non-empty `jobId`

### Requirement: A forge client supplies a job's log tail or says why not

The system SHALL fetch a GitHub job's log through its redirect and return the
last requested lines, and SHALL report one of unsupported, expired, forbidden
or unreachable when a forge can't supply a log.

#### Scenario: GitHub's redirect

- **WHEN** GitHub answers the log request with a redirect to a download
- **THEN** the client follows it and returns the last lines, oldest first,
  with `Truncated` set when there were more

#### Scenario: A log that is gone or off limits

- **WHEN** GitHub answers 404 or 410, or 401 or 403
- **THEN** the client reports expired, or forbidden

#### Scenario: Forgejo before v16

- **WHEN** a Forgejo job's log is requested
- **THEN** the client reports unsupported without making a request

### Requirement: Log requests are checked

The system SHALL reject a `lines` value outside 1 to 1000 with `400`, treat a
job ID that belongs to a different run as unknown, and require a session.

#### Scenario: A bad line count

- **WHEN** `lines` is 0, 1001, negative or not a number
- **THEN** the answer is `400`

#### Scenario: A job from another run

- **WHEN** the job ID exists but under a different run
- **THEN** the answer is `404`

### Requirement: The job log is readable through MCP

The system SHALL expose a read-only `get_job_log` tool returning what
`GET /api/runs/{runId}/jobs/{jobId}/log` returns, and SHALL report an unknown
job or an out-of-range `lines` as a tool error.

#### Scenario: A job's log

- **WHEN** a client calls `get_job_log` with a run and job ID
- **THEN** the result carries `available`, `lines`, `truncated` and `forgeUrl`

#### Scenario: An unknown job through MCP

- **WHEN** the job doesn't exist
- **THEN** the call is a tool error

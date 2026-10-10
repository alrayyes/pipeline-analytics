# http-observability Specification

## Purpose
Gives every HTTP request handled by the dashboard's server a
structured log record, so what's actually hitting the server —
including rejected and failed requests — is visible at the default
log level instead of only at start-up or on a hard internal error.

## Requirements

### Requirement: Every request is logged

The system SHALL emit one structured log record for every HTTP request it
handles, except requests to the liveness and readiness endpoints, recording
at minimum the method, path, response status, and duration.

#### Scenario: A successful request is logged

- **WHEN** a request completes with a 2xx or 3xx status
- **THEN** the system emits a log record for it at `Info` level

#### Scenario: A rejected request is logged

- **WHEN** a request is rejected before reaching its handler (e.g. no
  valid session)
- **THEN** the system still emits a log record for it, including the
  response status the rejection produced

#### Scenario: Probe requests are not logged

- **WHEN** a request is made to the liveness or the readiness endpoint and
  succeeds
- **THEN** the system emits no request log record for it

### Requirement: Log level reflects response outcome

The system SHALL log a request at a level determined by its response
status: `Error` for a 5xx status, `Warn` for a 4xx status, and `Info`
otherwise.

#### Scenario: A server error is logged at Error level

- **WHEN** a request's handler returns a 5xx status
- **THEN** the system logs that request at `Error` level

#### Scenario: A client error is logged at Warn level

- **WHEN** a request's handler returns a 4xx status
- **THEN** the system logs that request at `Warn` level

### Requirement: No sensitive data in request logs

The system SHALL NOT include the `Cookie` header, the `Authorization`
header, the request's raw query string, or any request or response
body content in a request log record.

#### Scenario: A request carrying a session cookie is logged safely

- **WHEN** a request that includes a session cookie is logged
- **THEN** the log record contains no cookie value

### Requirement: Readiness endpoint

The system SHALL expose an unauthenticated readiness endpoint that reports
whether the instance can serve, answering success only when its database
answers within a short deadline, and SHALL keep the liveness endpoint
independent of that check.

#### Scenario: Ready while the database answers

- **WHEN** the readiness endpoint is requested and the database answers
- **THEN** the system responds 200

#### Scenario: Not ready when the database does not answer

- **WHEN** the readiness endpoint is requested and the database errors or
  does not answer in time
- **THEN** the system responds 503 with a generic error body that names no
  internal detail, and logs the cause at `Error` level

#### Scenario: Liveness is independent of readiness

- **WHEN** the database is unreachable
- **THEN** the liveness endpoint still responds 200

#### Scenario: Container health follows readiness

- **WHEN** the container's health check runs
- **THEN** it reports healthy only if the readiness endpoint answers 200

### Requirement: Readiness is cached and drains on shutdown

The system SHALL reuse a readiness result for a few seconds so polling cannot
hammer the database, and on a shutdown signal SHALL answer 503 from `/readyz`
immediately while it keeps serving for a drain period, before it closes.

#### Scenario: Polling reuses the result

- **WHEN** `/readyz` is polled repeatedly inside the cache window
- **THEN** the database is probed once and every poll gets that result

#### Scenario: A shutdown signal takes the instance out of rotation first

- **WHEN** the process receives SIGTERM
- **THEN** `/readyz` answers 503 at once, other requests are still served for
  the drain period, and only then does the listener close

#### Scenario: Shutdown cannot hang

- **WHEN** a request never finishes
- **THEN** shutdown gives up after the shutdown timeout

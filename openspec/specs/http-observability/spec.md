# http-observability Specification

## Purpose
Gives every HTTP request handled by the dashboard's server a
structured log record, so what's actually hitting the server —
including rejected and failed requests — is visible at the default
log level instead of only at start-up or on a hard internal error.

## Requirements

### Requirement: Every request is logged

The system SHALL emit one structured log record for every HTTP
request it handles, except requests to the liveness-check endpoint,
recording at minimum the method, path, response status, and duration.

#### Scenario: A successful request is logged

- **WHEN** a request completes with a 2xx or 3xx status
- **THEN** the system emits a log record for it at `Info` level

#### Scenario: A rejected request is logged

- **WHEN** a request is rejected before reaching its handler (e.g. no
  valid session)
- **THEN** the system still emits a log record for it, including the
  response status the rejection produced

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

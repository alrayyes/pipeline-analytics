# Spec Delta

## MODIFIED Requirements

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

## ADDED Requirements

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

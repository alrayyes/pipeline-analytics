# Spec Delta

## ADDED Requirements

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

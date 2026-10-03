# Spec Delta

## ADDED Requirements

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

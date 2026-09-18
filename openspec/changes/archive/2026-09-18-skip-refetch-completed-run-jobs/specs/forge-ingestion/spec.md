# Spec Delta

## MODIFIED Requirements

### Requirement: Reconciliation polling

The system SHALL poll each tracked repository's recent workflow runs
at least hourly, using conditional requests (ETag/If-None-Match on
GitHub) to avoid rate-limit cost when nothing has changed, to backfill
history and catch missed webhook deliveries.

#### Scenario: Reconciliation backfills a newly tracked repo

- **WHEN** a repository is registered for tracking
- **THEN** the next reconciliation poll fetches and stores its recent
  workflow run history, not only future webhook events

#### Scenario: Unchanged repo costs no rate-limit budget

- **WHEN** a reconciliation poll runs against a GitHub repository with
  no new activity since the last poll
- **THEN** the system issues a conditional request and, on receiving a
  304 response, does not modify stored data and does not count the
  request against the token's primary rate limit

#### Scenario: Reconciliation catches a missed webhook

- **WHEN** a workflow run completed on the forge but no corresponding
  webhook delivery was ingested
- **THEN** the next reconciliation poll detects the discrepancy and
  stores the missing run, job, and step data

#### Scenario: An already-completed run's jobs aren't refetched

- **WHEN** a reconciliation poll's run listing reports a run as
  `completed` with the same conclusion already stored for it
- **THEN** the system does not issue a request for that run's jobs,
  even if other runs in the same poll did change

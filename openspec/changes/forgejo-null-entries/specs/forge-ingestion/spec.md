# Spec Delta

## ADDED Requirements

### Requirement: A malformed Forgejo response doesn't crash ingestion

The system SHALL skip a null run, job or step in a Forgejo response and keep
the entries around it, and SHALL NOT panic on any response body.

#### Scenario: A null run

- **WHEN** the runs response contains `null` in place of a run
- **THEN** the poll succeeds and returns the other runs

#### Scenario: A null job or step

- **WHEN** a jobs response contains `null` in place of a job, or a job's steps
  contain `null`
- **THEN** that entry is skipped and the rest are kept

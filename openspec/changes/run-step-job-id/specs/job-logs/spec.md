# Spec Delta

## ADDED Requirements

### Requirement: A step names its job

The system SHALL include the id of the job a step ran in on every step it
returns, so a client can ask for that job's log.

#### Scenario: A run's steps

- **WHEN** a client reads a run's steps
- **THEN** each step carries a non-empty `jobId`

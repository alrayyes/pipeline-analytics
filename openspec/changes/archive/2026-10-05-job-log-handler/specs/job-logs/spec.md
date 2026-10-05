# Spec Delta

## ADDED Requirements

### Requirement: The job log endpoint validates and scopes its request

The system SHALL reject a `lines` value outside 1 to 1000 with `400`, treat a
job id that belongs to a different run as unknown, and require a session.

#### Scenario: A bad line count

- **WHEN** `lines` is 0, 1001, negative or not a number
- **THEN** the answer is `400`

#### Scenario: A job from another run

- **WHEN** the job id exists but under a different run
- **THEN** the answer is `404`

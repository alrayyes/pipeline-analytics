# Spec Delta

## Purpose

Lets a signed-in user mark a flaky step as known, so it stops making its
pipeline look unhealthy, without hiding it or changing any measured figure.

## ADDED Requirements

### Requirement: A flaky step can be quarantined and un-quarantined

The system SHALL let a signed-in session quarantine a step, a step being one
name within one pipeline, with an optional note of at most 500 characters, and
un-quarantine it. Quarantining an already quarantined step SHALL renew the
mark: its 30 days restart and its note is replaced. The system SHALL answer
`404` for an unknown pipeline and `422` for a note over the limit. No forge
SHALL be contacted for either action.

#### Scenario: Quarantining a step

- **WHEN** a session quarantines the step `test` of a tracked pipeline with the
  note "waits on the shared database"
- **THEN** the answer is `200` and the step reports `quarantined: true` with
  that note and the time it was marked

#### Scenario: Un-quarantining a step

- **WHEN** a session un-quarantines a quarantined step
- **THEN** the step reports `quarantined: false` and raises the flaky-step
  health signal again if it is still flaky

#### Scenario: Renewing a quarantine

- **WHEN** a session quarantines a step that is already quarantined
- **THEN** its expiry is 30 days from now and the new note replaces the old

#### Scenario: The forge is not involved

- **WHEN** a session quarantines or un-quarantines a step
- **THEN** no request is made to the pipeline's forge

### Requirement: A quarantine mutes the health signal and nothing else

The system SHALL NOT count a quarantined step toward a pipeline's flaky-step
health signal or toward the flaky-step ratio in the failure insights. The
system SHALL leave every other figure unchanged: the step's failure rate, run
count and result matrix, and the failure rate and durations of the pipeline's
runs, and SHALL keep listing the step as flaky.

#### Scenario: A known flake no longer makes the pipeline unhealthy

- **WHEN** the only reason a pipeline was unhealthy is a flaky step, and that
  step is quarantined
- **THEN** the pipeline's health status is healthy and its signals no longer
  include the flaky-step signal

#### Scenario: Failures still count

- **WHEN** a pipeline's overall failure rate is over the unhealthy threshold
  and its flaky step is quarantined
- **THEN** the pipeline is still unhealthy, with the failure-rate signal

#### Scenario: The step stays in the flaky list

- **WHEN** a session lists flaky steps
- **THEN** a quarantined step is listed at its normal rank with the same flake
  rate, run count and matrix as before it was quarantined

#### Scenario: The ratio leaves it out of the numerator only

- **WHEN** one of a pipeline's four steps is flaky and quarantined
- **THEN** the flaky-step ratio for that pipeline is 0 and the step still
  counts among its four

### Requirement: A quarantine expires after 30 days

The system SHALL treat a quarantine as active for 30 days from the time it was
marked or last renewed, and SHALL treat an expired quarantine exactly as if it
had never been set, without any action from the user. The system SHALL report
when an active quarantine expires.

#### Scenario: An expired quarantine stops muting

- **WHEN** a step was quarantined 31 days ago and is still flaky
- **THEN** it reports `quarantined: false` and raises the flaky-step health
  signal

#### Scenario: Expiry is visible

- **WHEN** a session lists a quarantined step
- **THEN** the step reports the time its quarantine expires

### Requirement: Only a session may write a quarantine

The system SHALL accept quarantine and un-quarantine only from a signed-in
session, and SHALL answer `401` to an API token. Reading `quarantined`, the
note and the expiry SHALL be available to a session, an API token and the MCP
read tools, and the MCP server SHALL offer no tool that writes one.

#### Scenario: A token cannot quarantine

- **WHEN** a request with an API token tries to quarantine a step
- **THEN** the answer is `401` and nothing is stored

#### Scenario: An agent can see the mark

- **WHEN** an MCP client lists flaky steps
- **THEN** each reports `quarantined`, its note and its expiry

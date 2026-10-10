# Spec Delta

## ADDED Requirements

### Requirement: Run and step outcome

The system SHALL report each run and each step with a normalized outcome
(passed, failed, running, queued, cancelled, skipped or unknown) computed
from the forge's status and conclusion, so that a client does not have to
interpret forge status strings.

#### Scenario: A timed-out step is failed

- **WHEN** a step's conclusion is `timed_out`
- **THEN** its outcome is failed

#### Scenario: A conclusion wins over a stale status

- **WHEN** a run's status is still `in_progress` but its conclusion is
  `success`
- **THEN** its outcome is passed

#### Scenario: Pending work is running or queued

- **WHEN** a run has no conclusion and its status is `in_progress`
- **THEN** its outcome is running
- **AND WHEN** its status is `queued`
- **THEN** its outcome is queued

#### Scenario: An unrecognised state is never a pass

- **WHEN** a run has a conclusion the system does not recognise
- **THEN** its outcome is unknown, not passed

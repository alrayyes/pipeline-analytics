# Spec Delta

## ADDED Requirements

### Requirement: A Codecov upload error doesn't fail the build

The system SHALL report a failed Codecov upload as a warning and SHALL NOT fail
the job over it, so that an outage on Codecov's side doesn't turn main red. The
test and coverage steps keep failing the job when the code is wrong.

#### Scenario: Codecov refuses an upload

- **WHEN** Codecov answers an upload with a 403 or a 5xx
- **THEN** the job still passes, and the log shows the error

#### Scenario: A test fails while Codecov is down

- **WHEN** a test fails
- **THEN** the job fails at the test step, and the test results still upload

# Spec Delta

## ADDED Requirements

### Requirement: Telemetry window setting

The system SHALL persist, per account, the window the telemetry views
cover (24h, 7d or 30d), defaulting to 7d, so that the choice survives a
reload and carries between devices.

#### Scenario: Default window

- **WHEN** an account has never chosen a telemetry window
- **THEN** the system returns 7d

#### Scenario: Window choice is visible on the next read

- **WHEN** the user chooses 30d
- **THEN** a later settings read, from any session, returns 30d

#### Scenario: Invalid window is rejected

- **WHEN** an update sets the telemetry window to a value outside 24h,
  7d and 30d
- **THEN** the system rejects the update and leaves the stored window
  unchanged

#### Scenario: Resetting Pipelines filters leaves the window alone

- **WHEN** the user resets the Pipelines page's filters
- **THEN** the stored telemetry window is unchanged

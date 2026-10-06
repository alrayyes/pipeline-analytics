# account-settings Specification

## Purpose
Persists the dashboard account's UI preferences — theme and the
Pipelines/Repos page filters — server-side, so they follow the account
to any device or browser it logs in from, rather than resetting per
browser.

## Requirements

### Requirement: Settings persisted per account

The system SHALL persist, per dashboard account, the current value of:
theme (`light`/`dark`/`system`), forge filter, Pipelines health-status
filter, Pipelines repo selector, and Pipelines sort order.

#### Scenario: Setting survives a new session

- **WHEN** the user changes a setting and later starts a new
  authenticated session, on the same or a different device
- **THEN** the system returns the previously saved value for that
  setting

### Requirement: Settings defaults

The system SHALL return a documented default for any setting that has
never been explicitly set for the account.

#### Scenario: First-ever session has defaults

- **WHEN** an authenticated session requests settings before any
  setting has ever been saved for the account
- **THEN** the system returns the documented default for every setting

### Requirement: Settings update

The system SHALL allow an authenticated session to update one or more
settings in a single request, persisting each accepted value
immediately.

#### Scenario: Update is visible on next read

- **WHEN** the user updates a setting
- **THEN** a subsequent settings read, from any session, returns the
  updated value

#### Scenario: Invalid value is rejected

- **WHEN** an update request sets a setting to a value outside its
  documented set (e.g. an unrecognized theme)
- **THEN** the system rejects the update and leaves the previously
  stored value unchanged

### Requirement: Reset Pipelines filters to defaults

The system SHALL allow resetting the Pipelines page's filter-related
settings (health-status filter, repo selector, sort order) to their
documented defaults in a single request.

#### Scenario: Reset restores documented defaults

- **WHEN** the user resets the Pipelines page's filters
- **THEN** the system persists the documented default for the
  health-status filter, repo selector, and sort order

### Requirement: The settings response carries the defaults

The system SHALL return, with every settings response, the documented default
of each setting, whatever the account has stored.

#### Scenario: A stored value differs from its default

- **WHEN** an account has set the Pipelines health filter to `all`
- **THEN** the response's `pipelinesHealthFilter` is `all` and
  `defaults.pipelinesHealthFilter` is `unhealthy`

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

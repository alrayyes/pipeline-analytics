# Spec Delta

## ADDED Requirements

### Requirement: API tokens section on Settings

The system SHALL present, in a section of the Settings page, a form where
the signed-in user creates an API token by choosing its time-to-live from
presets of 30 days, 90 days and 1 year, showing the resulting expiry date
and updating it as the choice changes, and SHALL show the new token's raw
value once, on success. The section SHALL NOT list or revoke existing
tokens, and SHALL pass an automated axe-core scan.

#### Scenario: Default preset

- **WHEN** the user opens the API tokens section
- **THEN** the 90-day preset is selected, not the longest

#### Scenario: Expiry preview follows the choice

- **WHEN** the user changes the selected preset
- **THEN** the displayed expiry date updates to match, before the token
  is created

#### Scenario: No non-expiring option

- **WHEN** the user views the presets
- **THEN** none of them is a non-expiring token

#### Scenario: The token is shown once

- **WHEN** the user creates a token
- **THEN** its raw value is displayed once, with its expiry, and is not
  shown again after the user leaves or reloads the page

# Spec Delta

## ADDED Requirements

### Requirement: A tool can update the account's settings

The system SHALL expose as an MCP tool the partial update of the signed-in
account's settings, with the endpoint's meaning: a setting given a value is set,
one given null is cleared to its default, and one left out is unchanged.

#### Scenario: Changing one setting

- **WHEN** a client calls `update_settings` with only a theme
- **THEN** the theme changes, the other settings keep their values, and the
  result is the settings after the change

#### Scenario: Clearing a setting

- **WHEN** a client sets a setting to null
- **THEN** it returns to its documented default

#### Scenario: A value or key the settings do not allow

- **WHEN** a client gives a value outside a setting's set, or a key that is not
  a setting
- **THEN** the call is refused and nothing is stored

#### Scenario: An API token

- **WHEN** a client authenticated by an API token calls `update_settings`
- **THEN** the call succeeds for that token's account, as the endpoint does

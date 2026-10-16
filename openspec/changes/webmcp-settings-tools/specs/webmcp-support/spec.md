# Spec Delta

## ADDED Requirements

### Requirement: Tools read and change the account's settings

The system SHALL expose as WebMCP tools reading the signed-in account's
dashboard settings and changing them, the latter sending the same request as
the page's own controls, with a null clearing a setting to its default and the
server deciding what is valid. A change SHALL show on the open page without a
reload.

#### Scenario: Reading settings

- **WHEN** an in-browser agent calls `get_settings`
- **THEN** every setting and its default are returned, as the page's settings
  request answers them

#### Scenario: Changing and clearing settings

- **WHEN** an agent calls `update_settings` with a theme and a null forge filter
- **THEN** one PATCH carries both keys, the theme is set, the forge filter
  returns to its default, and the settings after the change are the result

#### Scenario: The open page follows

- **WHEN** `update_settings` succeeds
- **THEN** the page reloads its settings, so a changed theme or filter shows
  without a manual reload

#### Scenario: The person is not asked

- **WHEN** an agent calls `update_settings`
- **THEN** no confirmation is requested, as the page's controls ask for none

#### Scenario: The server refuses

- **WHEN** the server answers 400 for an invalid key or value
- **THEN** the call fails carrying the status and the server's message, nothing
  is changed and the page is not reloaded

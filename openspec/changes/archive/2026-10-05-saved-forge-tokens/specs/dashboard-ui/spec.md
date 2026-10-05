# Spec Delta

## ADDED Requirements

### Requirement: Saved forge tokens are managed in settings and used to register

The system SHALL list the saved forge tokens in settings, masked, and let the
user save one (replacing the token for that forge and instance) and delete one.
The register dialog SHALL say a saved token is in use and send no token when
one is saved for the chosen forge and instance, and SHALL ask for a token when
none is. The browser SHALL NOT keep a token: any left from before is removed.

#### Scenario: Save a token in settings

- **WHEN** the user saves a GitHub token in settings
- **THEN** settings lists it by its last four characters, and the token is
  not shown again anywhere

#### Scenario: Register with a saved token

- **WHEN** the user opens "Register repository" for a forge and instance that
  has a saved token, including after the browser's storage was cleared
- **THEN** the dialog says the saved token is in use, shows no token field,
  and registration and discovery are sent without a token

#### Scenario: No saved token

- **WHEN** the user deletes the saved token and opens "Register repository"
- **THEN** the dialog asks for a token

#### Scenario: A saved token the forge refuses

- **WHEN** discovery with the saved token fails
- **THEN** the failure is shown, the token field appears for a different
  token, and the saved token stays until it is replaced or deleted

#### Scenario: Tokens the browser remembered before

- **WHEN** the dashboard loads with tokens left in the browser's storage
- **THEN** they are removed and never offered

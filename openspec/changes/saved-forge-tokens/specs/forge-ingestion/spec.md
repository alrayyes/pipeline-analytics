# Spec Delta

## ADDED Requirements

### Requirement: A forge token can be saved and reused

The system SHALL let a signed-in session save one forge token per forge and
Forgejo instance, store it encrypted, return only its masked form, and use it
for registration and discovery when the request carries no token.

#### Scenario: Registering with the saved token

- **WHEN** a repo is registered with no token and a token is saved for its
  forge and instance
- **THEN** the saved token is used, and no response contains it

#### Scenario: No saved token

- **WHEN** a request carries no token and none is saved
- **THEN** the answer is `400` with code `no_saved_token`

#### Scenario: Replacing and deleting

- **WHEN** a token is saved for a forge and instance that already has one
- **THEN** it replaces it, and deleting it leaves already registered repos
  working

#### Scenario: An API token

- **WHEN** an API token, not a session, calls the saved token endpoints
- **THEN** the answer is `401`

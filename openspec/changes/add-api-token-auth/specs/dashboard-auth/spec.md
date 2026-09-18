# Spec Delta

## MODIFIED Requirements

### Requirement: Session-gated access

The system SHALL require a valid authenticated session or a valid API
token for every dashboard page and API endpoint except the
login/registration ceremony itself, the forge webhook receiver, and
token issuance/revocation, which require a session specifically.

#### Scenario: Unauthenticated request is denied

- **WHEN** a request to a dashboard page or data API arrives without a
  valid session
- **THEN** the system denies the request and redirects to the login
  ceremony

#### Scenario: Valid bearer token authenticates like a session

- **WHEN** a request to a data API carries `Authorization: Bearer
  <token>` for a token that is valid and unexpired
- **THEN** the system authenticates the request the same way a valid
  session would

#### Scenario: Revoked or expired token is denied

- **WHEN** a request carries a bearer token that has been revoked or
  has expired
- **THEN** the system returns `401` with the standard `Error` body,
  the same as an invalid session

## ADDED Requirements

### Requirement: API token issuance and revocation

The system SHALL let the authenticated user issue a long-lived,
revocable API token and later revoke it, both only via an active
session — not via another API token.

#### Scenario: Issuing a token

- **WHEN** the user calls `POST /api/auth/tokens` with a valid session
- **THEN** the system creates a new API token, returns its raw value
  once, and never returns that raw value again

#### Scenario: Revoking a token

- **WHEN** the user calls `DELETE /api/auth/tokens/{tokenId}` with a
  valid session, naming a token that belongs to them
- **THEN** the system revokes it, and any subsequent request bearing
  that token is denied

#### Scenario: A bearer token cannot manage tokens

- **WHEN** a request to `POST /api/auth/tokens` or `DELETE
  /api/auth/tokens/{tokenId}` carries only a bearer token, no session
- **THEN** the system denies the request

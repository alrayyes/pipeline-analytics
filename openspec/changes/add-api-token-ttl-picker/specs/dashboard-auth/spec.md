# Spec Delta

## MODIFIED Requirements

### Requirement: API token issuance and revocation

The system SHALL let the authenticated user issue a revocable API token
with a client-selectable time-to-live and later revoke it, both only via
an active session — not via another API token. The system SHALL clamp
the requested time-to-live to a fixed ceiling of 365 days regardless of
what the client requests, SHALL reject a requested time-to-live of zero
or less, and SHALL NOT offer a non-expiring token under any requested
value. The response to issuing a token SHALL carry the token's expiry.

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

#### Scenario: Requested TTL within the ceiling is honored

- **WHEN** the user calls `POST /api/auth/tokens` with `ttlSeconds` at or
  below the ceiling
- **THEN** the issued token expires `ttlSeconds` from now and the
  response's `expiresAt` says so

#### Scenario: Requested TTL above the ceiling is clamped

- **WHEN** the user calls `POST /api/auth/tokens` with `ttlSeconds` above
  the ceiling
- **THEN** the issued token expires at the ceiling, and the response's
  `expiresAt` shows the ceiling, not the requested value

#### Scenario: Omitted TTL falls back to the default

- **WHEN** the user calls `POST /api/auth/tokens` with no body, or with
  no `ttlSeconds`
- **THEN** the system issues the token with the default time-to-live of
  90 days, not the ceiling

#### Scenario: A zero or negative TTL is rejected

- **WHEN** the user calls `POST /api/auth/tokens` with `ttlSeconds` of
  zero or less
- **THEN** the system answers `400` and issues no token

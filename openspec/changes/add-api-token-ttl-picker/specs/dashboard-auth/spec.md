# Spec Delta

<!--
Sequencing note (see design.md "Sequencing" for the full explanation):
`add-api-token-auth` (#178/PR #186) is still unmerged, so the
requirement below does not exist in openspec/specs/dashboard-auth yet —
only in that change's own pending delta. This is written as ADDED,
combining that change's "API token issuance and revocation" requirement
with the TTL behavior this change adds, because MODIFIED requires the
requirement to already be present in the archived main spec. Once #186
archives, reconcile: drop the issuance/revocation scenarios this delta
duplicates from that change and re-file this delta as MODIFIED against
the now-real main-spec requirement, keeping only the TTL-specific
scenarios as the actual change.
-->

## ADDED Requirements

### Requirement: API token issuance and revocation

The system SHALL let the authenticated user issue a long-lived,
revocable API token with a client-selectable time-to-live and later
revoke it, both only via an active session — not via another API token.
The system SHALL clamp the requested time-to-live to a fixed ceiling
regardless of what the client requests, and SHALL NOT offer a
non-expiring token under any requested value.

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

- **WHEN** the user calls `POST /api/auth/tokens` with a requested
  time-to-live at or below the system's ceiling
- **THEN** the issued token's expiry reflects the requested value, not
  the ceiling

#### Scenario: Requested TTL above the ceiling is clamped

- **WHEN** the user calls `POST /api/auth/tokens` with a requested
  time-to-live above the system's ceiling
- **THEN** the issued token's expiry reflects the ceiling, not the
  requested value

#### Scenario: Omitted TTL falls back to a default

- **WHEN** the user calls `POST /api/auth/tokens` without a requested
  time-to-live
- **THEN** the system issues the token with a documented default
  time-to-live, not the ceiling

### Requirement: Token creation UI

The system SHALL present a page where the authenticated user creates an
API token by selecting its time-to-live from a fixed set of presets
before issuance, showing the resulting absolute expiry date and updating
it live as the selection changes.

#### Scenario: Default preset on first load

- **WHEN** the user opens the token creation page
- **THEN** the 90-day preset is selected by default, not the longest
  available option

#### Scenario: Expiry preview updates with the selection

- **WHEN** the user changes the selected TTL preset
- **THEN** the displayed absolute expiry date updates to match, before
  the token is created

#### Scenario: No non-expiring option is offered

- **WHEN** the user views the available TTL presets
- **THEN** none of them represents a non-expiring token

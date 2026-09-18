# Spec Delta

## ADDED Requirements

### Requirement: Additional credential enrollment

The system SHALL allow an authenticated session to register an
additional WebAuthn credential bound to the same account, distinct
from the anonymous first-run registration ceremony.

#### Scenario: Authenticated user adds a second passkey

- **WHEN** an authenticated user completes a WebAuthn registration
  ceremony from within their session
- **THEN** the system stores the new credential against their existing
  account without creating a second account

#### Scenario: Unauthenticated enrollment is rejected

- **WHEN** a WebAuthn registration ceremony targeting the "add a
  credential" endpoint is attempted without a valid session
- **THEN** the system denies the request

### Requirement: Credential listing

The system SHALL let an authenticated session list every credential
registered to their account, each with its user-supplied label and
creation date.

#### Scenario: All registered credentials are visible

- **WHEN** an authenticated user with more than one registered
  credential requests their credential list
- **THEN** the system returns every registered credential, including
  ones not used for the current session

### Requirement: Credential revocation

The system SHALL let an authenticated session revoke any of their
registered credentials except the account's last remaining one.

#### Scenario: Revoking an unused credential

- **WHEN** an authenticated user revokes a credential other than the
  one that authenticated their current session
- **THEN** the system deletes that credential, and it can no longer be
  used to log in

#### Scenario: Revoking the last remaining credential is rejected

- **WHEN** an authenticated user attempts to revoke their only
  remaining registered credential
- **THEN** the system rejects the revocation, so the account can never
  be left with no way to log in

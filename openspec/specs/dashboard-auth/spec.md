# dashboard-auth Specification

## Purpose
Protects the dashboard behind WebAuthn passkey login for its single solo-developer user, so the tool is not exposed with a password-only login despite being reachable at a public URL.

## Requirements

### Requirement: WebAuthn credential registration
The system SHALL allow exactly one user account to register a WebAuthn credential (passkey) on first run, and SHALL NOT offer a password-based registration or login path.

#### Scenario: First-run registration
- **WHEN** the dashboard is accessed for the first time with no WebAuthn credential registered
- **THEN** the system presents a WebAuthn registration ceremony and, on success, creates the single user account bound to that credential

#### Scenario: Second account is rejected
- **WHEN** a WebAuthn registration is attempted after a user account already exists
- **THEN** the system rejects the registration

### Requirement: WebAuthn login
The system SHALL authenticate the user via a WebAuthn assertion ceremony against their registered credential.

#### Scenario: Successful login
- **WHEN** the user completes a WebAuthn assertion matching their registered credential
- **THEN** the system establishes an authenticated session

#### Scenario: Failed assertion is rejected
- **WHEN** a WebAuthn assertion does not verify against the registered credential
- **THEN** the system denies the login attempt and does not establish a session

### Requirement: Session-gated access
The system SHALL require a valid authenticated session for every dashboard page and API endpoint except the login/registration ceremony itself and the forge webhook receiver.

#### Scenario: Unauthenticated request is denied
- **WHEN** a request to a dashboard page or data API arrives without a valid session
- **THEN** the system denies the request and redirects to the login ceremony

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

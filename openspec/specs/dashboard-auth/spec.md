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

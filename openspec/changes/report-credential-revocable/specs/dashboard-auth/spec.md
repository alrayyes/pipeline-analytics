# Spec Delta

## ADDED Requirements

### Requirement: Credential list reports what can be revoked

The system SHALL report, for each registered credential, whether it can be
revoked, so that a client does not apply the last-credential rule itself.

#### Scenario: Several credentials are all revocable

- **WHEN** an account has more than one credential and lists them
- **THEN** each is reported as revocable

#### Scenario: The last credential is not revocable

- **WHEN** an account has exactly one credential and lists it
- **THEN** it is reported as not revocable, and revoking it is still rejected

# Spec Delta

## ADDED Requirements

### Requirement: Registration rejects archived, forked, and mirror repos

The system SHALL verify, at registration time, that a repository is
not archived, is not a fork, and is not a mirror, and SHALL reject
registration with a reason if any of those is true.

#### Scenario: Registering a mirror is rejected

- **WHEN** the user submits a registration for a repository that is a
  mirror
- **THEN** the system rejects the registration and reports that it's a
  mirror, without creating a tracked record for it

#### Scenario: Registering a fork or archived repo is rejected

- **WHEN** the user submits a registration for a repository that is a
  fork or is archived
- **THEN** the system rejects the registration and reports the reason,
  without creating a tracked record for it

#### Scenario: The check is authoritative regardless of source

- **WHEN** a repository identifier reaches registration, whether from
  the discovery picker or typed manually
- **THEN** the system verifies its archived/fork/mirror status against
  the forge directly rather than trusting a prior discovery result

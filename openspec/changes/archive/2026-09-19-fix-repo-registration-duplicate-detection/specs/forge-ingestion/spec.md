# Spec Delta

## ADDED Requirements

### Requirement: Duplicate registration is rejected clearly

The system SHALL reject an attempt to register a repository identifier
that is already tracked on the same forge (and, for Forgejo, the same
instance URL) with a distinct response indicating it's already
tracked, rather than a generic internal error.

#### Scenario: Re-registering an already-tracked repo

- **WHEN** the user submits a registration for a repository identifier
  already tracked on that forge and instance
- **THEN** the system rejects it with a response identifying it as
  already tracked, and does not create a second record for it

#### Scenario: The registration UI never offers an already-tracked repo

- **WHEN** the user runs repository discovery
- **THEN** the system excludes every already-tracked repository on
  that forge and instance from the results, regardless of how many
  repositories are tracked in total

# Spec Delta

## ADDED Requirements

### Requirement: The untrack confirmation names the repository

The system SHALL name the repository by its identifier when the `untrack_repo`
tool asks the person to confirm, and SHALL name the id when the identifier cannot
be looked up, without refusing to ask.

#### Scenario: A listed repository

- **WHEN** the agent stops tracking a repository the list contains
- **THEN** the question names its identifier

#### Scenario: The lookup fails

- **WHEN** the list cannot be fetched
- **THEN** the question names the repository id and the person is still asked

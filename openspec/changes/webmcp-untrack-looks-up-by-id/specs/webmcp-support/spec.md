# Spec Delta

## ADDED Requirements

### Requirement: The untrack confirmation looks the repository up by id

The system SHALL find the identifier the `untrack_repo` confirmation names by
requesting that one repository by its id, not by reading the list, and SHALL
name the id when the lookup fails or finds nothing, without refusing to ask.

#### Scenario: A repository anywhere in the list

- **WHEN** the agent stops tracking a repository that is tracked
- **THEN** the question names its identifier, however many repositories are
  tracked

#### Scenario: An unknown id

- **WHEN** the agent stops tracking an id nothing is tracked under
- **THEN** the question names the id and the person is still asked

#### Scenario: The lookup is slow

- **WHEN** the lookup does not answer within three seconds
- **THEN** the question names the id and the person is still asked

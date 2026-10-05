# Spec Delta

## ADDED Requirements

### Requirement: A surviving frontend mutant fails the build

The system SHALL fail the frontend mutation job, and the matching pre-push
hook, when any mutant in the mutated modules survives, unless the mutant is
equivalent and the code says why.

#### Scenario: A test stops pinning a behaviour

- **WHEN** a change leaves a mutant in a mutated module surviving
- **THEN** the mutation job fails and the report names the mutant

#### Scenario: An equivalent mutant

- **WHEN** a mutant can't change behaviour
- **THEN** a Stryker disable comment beside it says why, and it doesn't fail
  the job

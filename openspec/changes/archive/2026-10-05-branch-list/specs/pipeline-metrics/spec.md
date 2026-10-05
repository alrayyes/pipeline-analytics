# Spec Delta

## ADDED Requirements

### Requirement: The branches with runs can be listed

The system SHALL list the branches that have runs started in the trailing
window, with their run counts, busiest first and then by name, and SHALL NOT
count a run with no recorded branch.

#### Scenario: A selector's options

- **WHEN** a client asks for the branches in the 7d window
- **THEN** each branch with a run in that window appears once with its count

#### Scenario: Runs with no branch

- **WHEN** some runs have no recorded branch
- **THEN** they aren't counted under any branch

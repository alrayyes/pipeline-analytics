# Spec Delta

## ADDED Requirements

### Requirement: Late-arriving content does not shift what is on screen

The system SHALL hold the space for content that loads after first paint, so
that its arrival does not move anything already in view.

#### Scenario: The passkey list on Settings

- **WHEN** the passkey rows load after the page renders
- **THEN** the controls below the table stay where they were

#### Scenario: The release history

- **WHEN** the release list is still loading
- **THEN** the footer is below the fold, and the list arriving does not move anything in view

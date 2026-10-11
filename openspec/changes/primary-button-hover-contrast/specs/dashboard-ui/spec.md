# Spec Delta

## ADDED Requirements

### Requirement: Interactive states keep AA colour contrast

The system SHALL keep a button's label at WCAG AA contrast (4.5:1 for normal
text) in its hover state as well as at rest.

#### Scenario: Hovering a primary button

- **WHEN** the pointer is over a default (primary) button
- **THEN** its label still meets 4.5:1 against the button's background

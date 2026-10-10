# Spec Delta

## ADDED Requirements

### Requirement: A control's visible text is part of its accessible name

The system SHALL give every control an accessible name that contains the text
it shows, and SHALL keep heading levels in order on each page.

#### Scenario: The time-window buttons

- **WHEN** the user views the time-window buttons on the failures or flaky page
- **THEN** each button's visible text ("24 hours", "7 days", "30 days") is its
  accessible name

#### Scenario: Settings headings

- **WHEN** a screen reader lists the Settings page's headings
- **THEN** no level is skipped

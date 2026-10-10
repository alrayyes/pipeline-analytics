# Spec Delta

## ADDED Requirements

### Requirement: Failure insights report their window

The system SHALL report the window the failure insights cover, and SHALL use
its own default and report that when the request names none or names one it
does not recognise, so a client does not need to know the default.

#### Scenario: A requested window is echoed

- **WHEN** failure insights are requested for a recognised window
- **THEN** the response names that window

#### Scenario: An omitted window reports the default

- **WHEN** failure insights are requested with no window
- **THEN** the response names the system's default window, whatever it is

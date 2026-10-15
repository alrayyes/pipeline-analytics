# Spec Delta

## MODIFIED Requirements

### Requirement: The release history renders in slices

The system SHALL render the release history a few releases at a time, so that
a long history does not hold the main thread in one long task, SHALL open on
the newest 20 releases, and SHALL reveal older ones, 20 at a time, on request
until every release has been shown.

#### Scenario: A long history

- **WHEN** the release history has more releases than fit in the first slice
- **THEN** the first releases appear at once and the rest of the first 20
  follow

#### Scenario: Older releases on request

- **WHEN** the history has more than 20 releases and the user chooses to show
  older releases
- **THEN** the next 20 releases appear below with keyboard focus on the first of
  them, and the control goes once every release is shown

#### Scenario: Lighthouse performance

- **WHEN** Lighthouse audits `/releases`
- **THEN** its performance score is at least 0.8

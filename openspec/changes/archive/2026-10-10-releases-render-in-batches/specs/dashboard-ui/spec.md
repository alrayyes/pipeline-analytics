# Spec Delta

## ADDED Requirements

### Requirement: The release history renders in slices

The system SHALL render the release history a few releases at a time, so that
a long history does not hold the main thread in one long task, and SHALL still
show every release.

#### Scenario: A long history

- **WHEN** the release history has more releases than fit in the first slice
- **THEN** the first releases appear at once and the rest follow, until every
  release is on the page

#### Scenario: Lighthouse performance

- **WHEN** Lighthouse audits `/releases`
- **THEN** its performance score is at least 0.8

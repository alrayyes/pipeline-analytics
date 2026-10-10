# Spec Delta

## ADDED Requirements

### Requirement: Quarantined steps are labelled

The system SHALL mark a quarantined step as "quarantined" wherever a flaky step
is shown, with the age of the mark, its note and when it expires, and SHALL let
a signed-in user quarantine or un-quarantine the step from there. The label
SHALL be readable without colour alone and SHALL pass an automated axe-core
scan.

#### Scenario: A quarantined step in the flaky list

- **WHEN** the user views the flaky list and one step is quarantined
- **THEN** that step shows a "quarantined" label, how long ago it was marked
  and when it expires, and offers "Un-quarantine"

#### Scenario: Quarantining from the list

- **WHEN** a signed-in user quarantines a flaky step from the list
- **THEN** the step shows the label without a page reload and its pipeline's
  health status updates

### Requirement: A quarantined step does not make its pipeline unhealthy

The system SHALL NOT show a pipeline as unhealthy, or list the flaky-step
signal on it, because of a flaky step that is quarantined, and SHALL NOT count
that step in the flaky-step ratio of the failure insights.

#### Scenario: Quarantine clears the flaky-step signal on the overview

- **WHEN** a pipeline's only unhealthy signal is a quarantined flaky step
- **THEN** the overview shows the pipeline healthy

#### Scenario: Unhealthy-steps overview

- **WHEN** a pipeline's only flaky step is quarantined
- **THEN** the cross-pipeline unhealthy-steps overview does not list that
  pipeline for it

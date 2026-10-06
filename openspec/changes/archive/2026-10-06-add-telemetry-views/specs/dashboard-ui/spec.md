# Spec Delta

## ADDED Requirements

### Requirement: Failure overview

The system SHALL present, for a selectable window (24h, 7d, 30d) and
an optional repo scope, the pass rate with its change against the
previous window, the failed-run count, the flaky-step ratio, the mean
time to recovery, the failure distribution by step, and the pipelines
failing most often.

#### Scenario: Window drives every figure

- **WHEN** the user changes the window from 7d to 30d
- **THEN** the system recomputes every figure on the page for the
  30-day window

#### Scenario: No recovery in the window

- **WHEN** no failed pipeline recovered within the window
- **THEN** the system shows the mean time to recovery as unavailable
  rather than zero

### Requirement: Runs list with stage progression

The system SHALL present a paginated list of runs filterable by
status (all, failed, running, success), each showing a progression bar
of its recorded steps, its status, its duration, its commit and
actor when known, and a link to the run on the originating forge.

#### Scenario: Failed step is visible in the progression bar

- **WHEN** a run has a failed step
- **THEN** the progression bar marks that step as failed, distinct
  from passed, running, skipped and not-yet-run steps, using more
  than colour alone

#### Scenario: Running runs refresh

- **WHEN** the list contains a running run
- **THEN** the system refreshes the list every 15 seconds until no
  listed run is running

#### Scenario: Filter change resets to the first page

- **WHEN** the user changes the status filter while past the first
  page
- **THEN** the system returns to the first page of the filtered
  results

### Requirement: Root cause diagnostics

The system SHALL present failures in the selected window grouped by
failing step, each group showing its occurrence count, affected
pipelines, a failure category, and a link to its failed runs.

#### Scenario: Group links to failed runs

- **WHEN** the user opens a failure group
- **THEN** the system links to the failed runs for that step, each
  with a deep link to the originating forge

#### Scenario: Unclassifiable failure

- **WHEN** a failure matches no category rule
- **THEN** the system lists it under "Uncategorised" rather than
  guessing a category

### Requirement: Flaky test telemetry

The system SHALL present each flaky step with its pass/fail result
across its most recent 40 runs, its flake rate and its run count,
ranked by flake rate.

#### Scenario: Matrix has a text equivalent

- **WHEN** the user views a step's result matrix
- **THEN** the system provides a text summary of its flakes and total
  runs readable without seeing the matrix

### Requirement: Phone-width layout and accessibility

The system SHALL render the four views without horizontal scroll at a
390px viewport width, and each SHALL pass an automated axe-core scan.

#### Scenario: View fits a phone

- **WHEN** any of the four views loads at 390px width
- **THEN** the page does not scroll horizontally

## MODIFIED Requirements

### Requirement: Flaky-step callout

The system SHALL visually distinguish a step flagged as flaky from a
step that is consistently failing or consistently passing, using a
treatment that does not rely on colour alone.

#### Scenario: Flaky step is distinguishable from a broken step

- **WHEN** a pipeline has both a flaky step and a consistently
  failing step
- **THEN** the system displays them with distinct visual treatment,
  so the user does not have to inspect run history to tell them apart

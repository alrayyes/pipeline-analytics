# dashboard-ui Specification

## Purpose
Presents ingested pipeline data and computed metrics visually, with drill-down and deep links to the originating GitHub/Forgejo run, so a solo developer can diagnose a pain point and act on it without the dashboard needing write access to either forge.

## Requirements

### Requirement: Repo/pipeline overview
The system SHALL present a list of tracked repositories and their pipelines with each pipeline's current health status.

#### Scenario: Unhealthy pipeline is visible at a glance
- **WHEN** the user opens the overview
- **THEN** the system displays every tracked pipeline's health status without requiring the user to open each pipeline individually

### Requirement: Pipeline detail view
The system SHALL present, for a selected pipeline, its duration trend (p50/p90) and failure-rate trend over a rolling window.

#### Scenario: Duration regression is visible in the detail view
- **WHEN** the user opens a pipeline whose duration trend has regressed
- **THEN** the system displays the trend chart showing the regression, not only the current aggregate value

### Requirement: Step-level breakdown
The system SHALL present, for a selected pipeline, its steps ranked by duration contribution, each step's queue-time vs. execution-time split, and each step's failure rate.

#### Scenario: Slowest step is identifiable in the UI
- **WHEN** the user opens a pipeline's step breakdown
- **THEN** the system displays the steps ordered by duration contribution, highest first

### Requirement: Flaky-step callout
The system SHALL visually distinguish a step flagged as flaky from a step that is consistently failing or consistently passing.

#### Scenario: Flaky step is distinguishable from a broken step
- **WHEN** a pipeline has both a flaky step and a consistently failing step
- **THEN** the system displays them with distinct visual treatment so the user does not have to inspect run history to tell them apart

### Requirement: Deep link to originating run
The system SHALL provide, for every displayed run, job, or step, a link to that item's page on the originating GitHub or Forgejo instance.

#### Scenario: User navigates to the failing run's log
- **WHEN** the user selects a failing step in the dashboard
- **THEN** the system provides a link that opens that step's log on the originating forge

### Requirement: Actions-minutes usage view
The system SHALL present runner-minutes usage per workflow for a tracked repository over a rolling window.

#### Scenario: Usage is visible per workflow
- **WHEN** the user opens a repository's usage view
- **THEN** the system displays runner minutes consumed broken down by workflow

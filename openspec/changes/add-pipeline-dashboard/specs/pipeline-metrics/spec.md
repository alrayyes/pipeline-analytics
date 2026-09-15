## Purpose

Derives health scores and pain-point signals from ingested pipeline run data — duration trends, failure rates, flakiness, slow steps, and usage cost — so a solo developer can tell at a glance whether a pipeline is healthy and where to focus effort.

## ADDED Requirements

### Requirement: Pipeline health status
The system SHALL compute a health status for each tracked pipeline, derived from its recent failure rate, duration trend, and flaky-step signal.

#### Scenario: Healthy pipeline
- **WHEN** a pipeline's recent runs show a low failure rate, no significant duration regression, and no flaky steps
- **THEN** the system reports that pipeline's health status as healthy

#### Scenario: Unhealthy pipeline
- **WHEN** a pipeline's recent runs show an elevated failure rate, a significant duration regression, or a flaky step
- **THEN** the system reports that pipeline's health status as unhealthy and identifies which signal(s) triggered it

### Requirement: Duration trend by percentile
The system SHALL compute p50 and p90 duration for each pipeline and for each step within a pipeline, over a rolling window of recent runs.

#### Scenario: Step duration regression is visible in the trend
- **WHEN** a step's p90 duration over the most recent runs is materially higher than its p90 duration over the prior window
- **THEN** the system's computed trend for that step reflects the increase

### Requirement: Queue time vs. execution time split
The system SHALL separately track, for each job, the time spent queued awaiting a runner and the time spent executing.

#### Scenario: Queue-bound job is distinguishable from a slow job
- **WHEN** a job's total duration is dominated by queue wait rather than step execution
- **THEN** the system reports queue time and execution time as separate values for that job, not combined into a single duration figure

### Requirement: Failure rate trend
The system SHALL compute a failure rate, over a rolling window of recent runs, for each pipeline and for each step within a pipeline.

#### Scenario: Step-level failure rate is available
- **WHEN** a specific step has failed in some but not all recent runs of its pipeline
- **THEN** the system reports that step's failure rate independent of the pipeline's overall failure rate

### Requirement: Flaky step detection
The system SHALL identify a step as flaky when its outcome is inconsistent (both passing and failing) across recent runs on unchanged or near-unchanged code, distinct from a step that fails consistently.

#### Scenario: Flaky step flagged separately from a broken step
- **WHEN** a step alternates between pass and fail across recent runs
- **THEN** the system flags that step as flaky, distinct from a step that has failed in every recent run

### Requirement: Slowest-step ranking
The system SHALL rank the steps within a pipeline by their contribution to total pipeline duration.

#### Scenario: Slowest step is identifiable
- **WHEN** a user requests the duration breakdown for a pipeline
- **THEN** the system returns that pipeline's steps ordered by their typical duration contribution, highest first

### Requirement: Actions-minutes usage tracking
The system SHALL track runner time consumed per workflow and per repository, over a rolling window.

#### Scenario: Usage is attributable to a workflow
- **WHEN** a user requests usage for a tracked repository
- **THEN** the system reports total runner minutes consumed, broken down by workflow, for the requested window

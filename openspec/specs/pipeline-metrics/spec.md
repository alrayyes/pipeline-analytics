# pipeline-metrics Specification

## Purpose
Derives health scores and pain-point signals from ingested pipeline run data — duration trends, failure rates, flakiness, slow steps, and usage cost — so a solo developer can tell at a glance whether a pipeline is healthy and where to focus effort.

## Requirements

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

### Requirement: Telemetry can be scoped to one branch

The system SHALL accept an optional `branch` on the failure insights, the run
list and the flaky steps, cover only runs on that exact branch when it is
given and every branch when it is not, and reject one longer than 255
characters.

#### Scenario: A healthy main beside a broken feature branch

- **WHEN** the insights are read with `branch=main`
- **THEN** only main's runs count toward the totals and the pass rate

#### Scenario: A branch with no runs

- **WHEN** a branch has no runs in the window
- **THEN** the answer is an empty result, not an error

#### Scenario: A branch that is too long

- **WHEN** `branch` is longer than 255 characters
- **THEN** the answer is `400`

### Requirement: The branches with runs can be listed

The system SHALL list the branches that have runs started in the trailing
window, with their run counts, busiest first and then by name, and SHALL NOT
count a run with no recorded branch.

#### Scenario: A selector's options

- **WHEN** a client asks for the branches in the 7d window
- **THEN** each branch with a run in that window appears once with its count

#### Scenario: Runs with no branch

- **WHEN** some runs have no recorded branch
- **THEN** they aren't counted under any branch

### Requirement: Window failure aggregates

The system SHALL compute, over a selectable window and optional repo
scope, total runs, failed runs, pass rate, pass-rate change versus the
preceding window of equal length, flaky-step ratio, and mean time to
recovery.

#### Scenario: Pass-rate change

- **WHEN** the window's pass rate is 78.4% and the preceding window's
  is 82.6%
- **THEN** the system reports a change of -4.2 percentage points

#### Scenario: Mean time to recovery

- **WHEN** a pipeline fails and its next run succeeds 40 minutes
  afterward
- **THEN** that recovery contributes 40 minutes to the mean

### Requirement: Failure grouping by step and category

The system SHALL group failed steps in the window by step name with
occurrence counts and affected pipelines, and assign each group one of
a fixed set of categories (infrastructure, code/tests, network/
timeouts, config/secrets) from the step's conclusion and name, or
`uncategorised` when no rule matches.

#### Scenario: Timeout categorized

- **WHEN** a step concluded `timed_out`
- **THEN** the system assigns it to network/timeouts

#### Scenario: No rule matches

- **WHEN** a failed step matches no category rule
- **THEN** the system assigns it to `uncategorised`

### Requirement: Flaky steps ranked with their recent results

The system SHALL list flaky steps across every pipeline for a window, ranked
by flake rate, each with its run count and its result in up to its 40 most
recent runs, oldest first.

#### Scenario: Flakiest first

- **WHEN** two steps are flagged flaky and one failed in a larger share of
  its runs in the window
- **THEN** that step is listed first

#### Scenario: A step with fewer than 40 runs

- **WHEN** a flaky step has run 12 times
- **THEN** its recent results hold those 12 outcomes, not 40

### Requirement: Failure shares

The system SHALL report, for the selected window, each failing step's share
and each failure category's share of all failed-step occurrences, so that a
client does not compute percentages from counts.

#### Scenario: Shares add up

- **WHEN** failed steps exist in the window
- **THEN** the shares of the failing steps sum to 1, and so do the shares of
  the failure categories

#### Scenario: Nothing failed

- **WHEN** no step failed in the window
- **THEN** both lists are empty and no share is reported as zero for
  something that did not happen

#### Scenario: Categories are ordered by weight

- **WHEN** the category breakdown is reported
- **THEN** it lists the heaviest category first

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

### Requirement: Run and step outcome

The system SHALL report each run and each step with a normalized outcome
(passed, failed, running, queued, cancelled, skipped or unknown) computed
from the forge's status and conclusion, so that a client does not have to
interpret forge status strings.

#### Scenario: A timed-out step is failed

- **WHEN** a step's conclusion is `timed_out`
- **THEN** its outcome is failed

#### Scenario: A conclusion wins over a stale status

- **WHEN** a run's status is still `in_progress` but its conclusion is
  `success`
- **THEN** its outcome is passed

#### Scenario: Pending work is running or queued

- **WHEN** a run has no conclusion and its status is `in_progress`
- **THEN** its outcome is running
- **AND WHEN** its status is `queued`
- **THEN** its outcome is queued

#### Scenario: An unrecognised state is never a pass

- **WHEN** a run has a conclusion the system does not recognise
- **THEN** its outcome is unknown, not passed

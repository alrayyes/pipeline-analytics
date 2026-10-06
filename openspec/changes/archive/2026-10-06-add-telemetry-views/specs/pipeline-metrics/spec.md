# Spec Delta

## ADDED Requirements

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
  later
- **THEN** that recovery contributes 40 minutes to the mean

### Requirement: Failure grouping by step and category

The system SHALL group failed steps in the window by step name with
occurrence counts and affected pipelines, and assign each group one of
a fixed set of categories (infrastructure, code/tests, network/
timeouts, config/secrets) from the step's conclusion and name, or
"uncategorised" when no rule matches.

#### Scenario: Timeout categorised

- **WHEN** a step concluded `timed_out`
- **THEN** the system assigns it to network/timeouts

#### Scenario: No rule matches

- **WHEN** a failed step matches no category rule
- **THEN** the system assigns it to uncategorised

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

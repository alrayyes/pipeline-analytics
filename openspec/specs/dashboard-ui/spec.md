# dashboard-ui Specification

## Purpose
Presents ingested pipeline data and computed metrics visually, with drill-down and deep links to the originating GitHub/Forgejo run, so a solo developer can diagnose a pain point and act on it without the dashboard needing write access to either forge.

## Requirements

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

The system SHALL visually distinguish a step flagged as flaky from a
step that is consistently failing or consistently passing, using a
treatment that does not rely on color alone.

#### Scenario: Flaky step is distinguishable from a broken step

- **WHEN** a pipeline has both a flaky step and a consistently
  failing step
- **THEN** the system displays them with distinct visual treatment,
  so the user does not have to inspect run history to tell them apart

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

### Requirement: Cross-pipeline unhealthy-steps overview
The system SHALL present every flaky or failing step across every
tracked pipeline, grouped by pipeline, paginated over pipeline groups
when the number of pipelines with an unhealthy step exceeds one page.

#### Scenario: Unhealthy step is visible without opening each pipeline
- **WHEN** the user opens the Steps overview
- **THEN** the system displays every tracked pipeline that has a flaky
  or failing step, grouped by pipeline, without requiring the user to
  open each pipeline individually

#### Scenario: Pipeline-group count exceeds one page
- **WHEN** more pipelines have an unhealthy step than fit on one page
- **THEN** the system renders only the current page's pipeline groups,
  with a control to advance to the next page and, once past the first
  page, back to the previous one

#### Scenario: Forge filter change resets to the first page
- **WHEN** the user is on a page other than the first and changes the
  forge filter
- **THEN** the system returns to the first page of the newly filtered
  results, rather than staying on an offset the new result set may not
  reach

### Requirement: Saved forge tokens are managed in settings and used to register

The system SHALL list the saved forge tokens in settings, masked, and let the
user save one (replacing the token for that forge and instance) and delete one.
The register dialog SHALL say a saved token is in use and send no token when
one is saved for the chosen forge and instance, and SHALL ask for a token when
none is. The browser SHALL NOT keep a token: any left from before is removed.

#### Scenario: Save a token in settings

- **WHEN** the user saves a GitHub token in settings
- **THEN** settings lists it by its last four characters, and the token is
  not shown again anywhere

#### Scenario: Register with a saved token

- **WHEN** the user opens "Register repository" for a forge and instance that
  has a saved token, including after the browser's storage was cleared
- **THEN** the dialog says the saved token is in use, shows no token field,
  and registration and discovery are sent without a token

#### Scenario: No saved token

- **WHEN** the user deletes the saved token and opens "Register repository"
- **THEN** the dialog asks for a token

#### Scenario: A saved token the forge refuses

- **WHEN** discovery with the saved token fails
- **THEN** the failure is shown, the token field appears for a different
  token, and the saved token stays until it is replaced or deleted

#### Scenario: Tokens the browser remembered before

- **WHEN** the dashboard loads with tokens left in the browser's storage
- **THEN** they are removed and never offered

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
  than color alone

#### Scenario: Running runs refresh

- **WHEN** the list contains a run that is still in progress
- **THEN** the system refreshes the list every 15 seconds until no
  listed run is running

#### Scenario: Status filter change returns the runs list to its first page

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
- **THEN** the system lists it under `Uncategorised` rather than
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

### Requirement: Footer links to the source and the license

The system SHALL link the footer to the project's GitHub repository and to its
license, on every page.

#### Scenario: The source link

- **WHEN** the footer is shown on any page
- **THEN** it has a "GitHub" link, with the GitHub mark, to the repository

#### Scenario: The license link

- **WHEN** the footer is shown on any page
- **THEN** it has an "AGPL-3.0" link to the repository's `LICENSE` file

#### Scenario: A narrow screen

- **WHEN** the footer is shown on a 360px-wide viewport
- **THEN** its links wrap onto further rows, the page does not scroll
  sideways, and no separator is left at the end of a row

### Requirement: Footer version links to the release history

The system SHALL show the running build's version in the footer and, for a
release build, link it to the release history page, without a separate
release-history link.

#### Scenario: The version opens the release history

- **WHEN** the footer shows a release version
- **THEN** the version is a link to the release history page

#### Scenario: No second link to the same page

- **WHEN** the footer is shown on any page
- **THEN** it has no separate release-history link

#### Scenario: No link to the page you are on

- **WHEN** the user is on the release history page
- **THEN** the footer shows the version as plain text

#### Scenario: A dev build has nothing to link

- **WHEN** the running build is a development build
- **THEN** the footer says so and shows no link

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

### Requirement: Pipeline overview list

The system SHALL present a list of tracked repositories and their
pipelines with each pipeline's current health status, paginated when the
number of tracked pipelines exceeds one page, with the health-status filter
and the sort order applied across every tracked pipeline.

#### Scenario: Unhealthy pipeline is visible at a glance

- **WHEN** the user opens the overview
- **THEN** the system displays every tracked pipeline's health status
  without requiring the user to open each pipeline individually

#### Scenario: Pipeline count exceeds one page

- **WHEN** more pipelines are tracked than fit on one page
- **THEN** the system renders only the current page's pipelines, with a
  control to advance to the next page and, once past the first page, back
  to the previous one

#### Scenario: Filter change resets to the first page

- **WHEN** the user is on a page other than the first and changes the
  health-status filter, repo selector, sort order, or forge filter
- **THEN** the system returns to the first page of the newly
  filtered/sorted results, rather than staying on an offset the new result
  set may not reach

#### Scenario: Health-status filter and "most recently run" sort apply across every pipeline

- **WHEN** the health-status filter is set to something other than "all",
  or the sort order is set to "most recently run"
- **THEN** the system applies that filter or sort across every tracked
  pipeline before paging, so a pipeline matching the filter on a later page
  appears on its page in order, every page but the last is full, and
  whether another page exists is accurate

#### Scenario: An unrecognised filter or sort is rejected

- **WHEN** a request names a health status or sort order the system does
  not know
- **THEN** the system rejects it instead of returning an empty list

### Requirement: Release history

The system SHALL present the project's release history, newest first, from
release notes bundled with the build, without querying a third-party service
when the page loads.

#### Scenario: History renders without a third-party request

- **WHEN** the user opens the release history page
- **THEN** the system lists the releases from data shipped with the
  application and makes no request to GitHub

#### Scenario: Each release links to its page on the forge

- **WHEN** the release history is shown
- **THEN** each release's title links to that release on GitHub

#### Scenario: History is unavailable

- **WHEN** the bundled release data can't be loaded
- **THEN** the system says so and links to the full list on GitHub, rather
  than showing a blank page

### Requirement: Telemetry views can be scoped to one branch

The system SHALL offer a branch selector on the overview, root-cause, flaky and
runs views, listing the branches that have runs in the selected window, and
SHALL send the chosen branch to each view's read. With no branch chosen the
views cover every branch.

#### Scenario: Pick a branch

- **WHEN** the user picks `main` in the selector
- **THEN** the overview, root-cause, flaky and runs views show only `main`'s
  runs, including the overview's latest failure, and the choice carries from
  one view to the next

#### Scenario: No selection

- **WHEN** the user has not picked a branch
- **THEN** every view covers every branch, and the selector reads "All
  branches"

#### Scenario: A branch with no runs

- **WHEN** the chosen branch has no runs in the window
- **THEN** the view says so and names the branch, rather than showing zeros

#### Scenario: Branches follow the window

- **WHEN** the user changes the time window
- **THEN** the selector offers the branches that have runs in the new window,
  and a chosen branch stays chosen even if it has none there

#### Scenario: The branch list cannot load

- **WHEN** the branch list request fails
- **THEN** the selector offers only "All branches" and the views still load

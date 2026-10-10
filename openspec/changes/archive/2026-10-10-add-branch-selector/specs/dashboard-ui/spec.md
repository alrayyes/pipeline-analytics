# Spec Delta

## ADDED Requirements

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

# Spec Delta

## ADDED Requirements

### Requirement: Tools cover the runs, logs, insights and flaky-step reads

The system SHALL expose as WebMCP tools a pipeline's steps, a job's log, the
failure insights, the run list, the flaky steps and the branch list, each
calling the same fetch function the dashboard page for it calls.

#### Scenario: Reading the run list

- **WHEN** an in-browser agent calls `list_runs` with a status and a branch
- **THEN** the dashboard's run fetch runs with that status and branch, and the
  result is the run list the Runs page shows

#### Scenario: A filter left out

- **WHEN** an agent calls `get_failure_insights` with no window
- **THEN** no window is sent and the server's own default applies

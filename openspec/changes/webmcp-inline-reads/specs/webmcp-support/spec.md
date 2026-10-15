# Spec Delta

## ADDED Requirements

### Requirement: Tools cover a run's steps, a step's flaky runs and the unhealthy steps

The system SHALL expose as WebMCP tools a run's steps, the runs in which a
pipeline's step flaked, and the pipelines with a flaky or failing step, each
calling the same fetch function the dashboard uses for it.

#### Scenario: Opening a run an agent found

- **WHEN** an in-browser agent calls `get_run_steps` with a run ID from
  `list_runs` or `list_pipeline_flaky_runs`
- **THEN** the run page's own fetch runs for that ID and the result is that
  run's steps

#### Scenario: A run that does not exist

- **WHEN** the run page is opened for a run the server does not know
- **THEN** it says the run was not found, as before the fetch was shared

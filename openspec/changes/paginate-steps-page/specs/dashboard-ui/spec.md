# Spec Delta

## ADDED Requirements

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

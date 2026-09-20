# Spec Delta

## MODIFIED Requirements

### Requirement: Repo/pipeline overview
The system SHALL present a list of tracked repositories and their
pipelines with each pipeline's current health status, paginated when the
number of tracked pipelines exceeds one page.

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

#### Scenario: Health-status filter and "most recently run" sort apply per page
- **WHEN** the health-status filter is set to something other than "all",
  or the sort order is set to "most recently run"
- **THEN** the system applies that filter or sort within the currently
  loaded page's pipelines only, not across every tracked pipeline -- a
  pipeline matching the filter on a later page isn't pulled forward, and a
  page can show fewer results than its size, or none, even when a
  matching pipeline exists elsewhere in the list

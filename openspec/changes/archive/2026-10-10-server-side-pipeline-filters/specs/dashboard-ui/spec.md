# Spec Delta

## REMOVED Requirements

### Requirement: Repo/pipeline overview

**Reason**: One of its scenarios documented that the health filter and the
"most recently run" sort applied only to the loaded page, which is no longer
true: both are now applied across every pipeline by the server. A modified
requirement can't drop a scenario, so the requirement is retired and
replaced.

**Migration**: Replaced by "Pipeline overview list", which carries the other
three scenarios unchanged and the corrected filter and sort behaviour.

## ADDED Requirements

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

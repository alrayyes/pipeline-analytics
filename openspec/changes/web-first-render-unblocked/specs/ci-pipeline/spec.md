# Spec Delta

## ADDED Requirements

### Requirement: Lighthouse gates render-blocking requests and request chains

The system SHALL fail the lighthouse job when a page has render-blocking
requests or a critical request chain, on `/login` and on each signed-in page.

#### Scenario: Stylesheet requested separately from the page

- **WHEN** the dashboard's stylesheet is served as its own render-blocking
  request
- **THEN** `render-blocking-insight` scores below 0.9 and the job fails

#### Scenario: A resource discovered behind another

- **WHEN** a critical resource such as the web font is only requested after
  the stylesheet has loaded
- **THEN** `network-dependency-tree-insight` scores below 0.9 and the job fails

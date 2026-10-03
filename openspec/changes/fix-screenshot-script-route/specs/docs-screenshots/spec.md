# Spec Delta

## ADDED Requirements

### Requirement: Screenshots are taken at the route the page lives at

The capture script SHALL load each page it captures at that page's own
route, so moving a page does not leave the script waiting for controls the
landing page never had.

#### Scenario: The Pipelines overview is captured at its route

- **WHEN** the script captures the Pipelines overview
- **THEN** it loads `/pipelines`, not the site root

#### Scenario: A moved page fails the capture run

- **WHEN** a page the script drives moves and the script isn't updated
- **THEN** the run fails and prints which control it waited for

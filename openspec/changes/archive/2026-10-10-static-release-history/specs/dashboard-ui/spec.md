# Spec Delta

## ADDED Requirements

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

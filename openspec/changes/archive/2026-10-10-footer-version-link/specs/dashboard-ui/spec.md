# Spec Delta

## ADDED Requirements

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

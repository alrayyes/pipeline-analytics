# Spec Delta

## ADDED Requirements

### Requirement: Footer links to the source and the license

The system SHALL link the footer to the project's GitHub repository and to its
license, on every page.

#### Scenario: The source link

- **WHEN** the footer is shown on any page
- **THEN** it has a "GitHub" link, with the GitHub mark, to the repository

#### Scenario: The license link

- **WHEN** the footer is shown on any page
- **THEN** it has an "AGPL-3.0" link to the repository's `LICENSE` file

#### Scenario: A narrow screen

- **WHEN** the footer is shown on a 360px-wide viewport
- **THEN** its links wrap onto further rows, the page does not scroll
  sideways, and no separator is left at the end of a row

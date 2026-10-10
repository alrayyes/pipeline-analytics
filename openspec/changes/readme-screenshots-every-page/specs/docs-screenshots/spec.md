# Spec Delta

## ADDED Requirements

### Requirement: Every page except the footer's has a README screenshot

The system SHALL capture one screenshot of each page under `web/src/routes/`,
from mocked data with no real repository names, tokens or user data, and the
README SHALL show it with alt text that says what the page is for. The
footer's own pages, `/releases` and `/legal`, SHALL have none.

#### Scenario: A page with no screenshot

- **WHEN** a page exists under `web/src/routes/` that is neither covered by a
  screenshot nor one of the footer's pages
- **THEN** the unit tests fail

#### Scenario: The footer's pages

- **WHEN** the README's screenshots are listed
- **THEN** none shows `/releases` or `/legal`

### Requirement: The README's screenshot section is generated

The system SHALL write the README's screenshot entries between markers from
the captured set, as part of the capture run, so that a page added later
appears in the README without a hand edit.

#### Scenario: A capture run

- **WHEN** `web/scripts/capture-screenshots.sh` completes
- **THEN** the README section between the markers lists every captured
  screenshot with its alt text

#### Scenario: The section drifted from the list

- **WHEN** the README section differs from the one generated from the list
- **THEN** the unit tests fail

### Requirement: A release proposes the screenshots in a pull request

The system SHALL, after a release that changed the UI, open a pull request
carrying the recaptured images and the generated README section, and SHALL NOT
push them to the default branch. It SHALL open one even when an earlier
release's pull request from the same branch has already merged.

#### Scenario: A release after a merged screenshot pull request

- **WHEN** the screenshots job runs and the previous screenshot pull request
  has merged
- **THEN** it opens a new pull request from the refreshed branch

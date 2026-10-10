# Spec Delta

## ADDED Requirements

### Requirement: Footer links that leave the app open in a new tab

The system SHALL open the footer's links to other sites (the GitHub repository
and the license) in a new tab, without giving the opened page access to the
dashboard, and SHALL tell assistive technology that they do. The footer's links
to the app's own pages SHALL open in the same tab.

#### Scenario: Leaving for the repository

- **WHEN** the user follows the footer's GitHub or license link
- **THEN** the page opens in a new tab and the dashboard stays open in the
  first, and the link's accessible name says it opens in a new tab

#### Scenario: The app's own footer pages

- **WHEN** the user follows the version or Privacy & disclaimer link
- **THEN** the page opens in the same tab

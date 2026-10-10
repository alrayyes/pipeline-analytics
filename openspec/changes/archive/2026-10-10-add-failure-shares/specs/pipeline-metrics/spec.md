# Spec Delta

## ADDED Requirements

### Requirement: Failure shares

The system SHALL report, for the selected window, each failing step's share
and each failure category's share of all failed-step occurrences, so that a
client does not compute percentages from counts.

#### Scenario: Shares add up

- **WHEN** failed steps exist in the window
- **THEN** the shares of the failing steps sum to 1, and so do the shares of
  the failure categories

#### Scenario: Nothing failed

- **WHEN** no step failed in the window
- **THEN** both lists are empty and no share is reported as zero for
  something that did not happen

#### Scenario: Categories are ordered by weight

- **WHEN** the category breakdown is reported
- **THEN** it lists the heaviest category first

# Spec Delta

## ADDED Requirements

### Requirement: Page wiring is tested without a browser

The system SHALL test what a page asks the server for and what it shows for
each state (nothing to show, a failed request, no saved setting) in a component
test that mounts the page in a DOM and fakes the network, and SHALL keep the
end-to-end suite to one journey per flow.

#### Scenario: A page's window toggle

- **WHEN** the Failures page is mounted and a different time window is chosen
- **THEN** the next request to the server carries that window and the choice is
  shown as selected, asserted in a component test

#### Scenario: A state behind a flow

- **WHEN** a page's empty, error or no-saved-setting state needs a test
- **THEN** it is a component test, not another end-to-end test of the same flow

#### Scenario: Plain unit tests keep their environment

- **WHEN** a test file does not import the DOM helper
- **THEN** it runs without a `window`

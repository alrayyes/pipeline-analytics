# Spec Delta

## ADDED Requirements

### Requirement: Test results reach Codecov for both suites

The system SHALL write JUnit results from the Go and frontend tests and upload
them to Codecov as test results, including when a test failed, except on
Dependabot runs, which have no upload token.

#### Scenario: A test fails

- **WHEN** a Go or frontend test fails
- **THEN** its JUnit results are still uploaded

#### Scenario: A Dependabot run

- **WHEN** Dependabot opens a pull request
- **THEN** no test results are uploaded and the job doesn't fail on a missing
  token

#### Scenario: The upload input is misspelled

- **WHEN** the workflow says `report-type` instead of `report_type`
- **THEN** the `changes` job fails

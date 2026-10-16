# Spec Delta

## ADDED Requirements

### Requirement: Tools read the version and GitHub rate-limit use

The system SHALL expose as WebMCP tools reading the running service's version
and the GitHub rate-limit use per token, each sending the same request as the
page that shows it.

#### Scenario: Reading the version

- **WHEN** an in-browser agent calls `get_version`
- **THEN** the footer's request is sent and the build's version is returned

#### Scenario: Reading rate-limit use

- **WHEN** an agent calls `get_github_rate_limit`
- **THEN** the Insights page's request is sent and the result lists each token,
  masked, with its repositories and its latest observed status when it has one

#### Scenario: The server refuses

- **WHEN** the server answers either request with an error
- **THEN** the call fails and carries the status

### Requirement: Every API operation is a tool or a written exclusion

The system SHALL fail its unit tests when an operation in `openapi.yaml` has
neither a WebMCP tool nor an `x-agent-excluded` reason, and SHALL keep no list
of operations awaiting a tool.

#### Scenario: A new operation with no tool

- **WHEN** an operation is added to the spec without a tool or an exclusion
- **THEN** the parity test fails naming it

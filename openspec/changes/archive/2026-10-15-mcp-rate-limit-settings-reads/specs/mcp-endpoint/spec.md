# Spec Delta

## ADDED Requirements

### Requirement: Tools cover the rate-limit status and the settings

The system SHALL expose as MCP tools the GitHub rate-limit status per token and
the signed-in account's settings, each calling the same code as its endpoint.

#### Scenario: Repos sharing a token

- **WHEN** an agent calls `get_github_rate_limit` and two repos use one token
- **THEN** the result holds one entry for that token, naming both repos, with
  its masked token and its last observed status

#### Scenario: A token not yet used

- **WHEN** a GitHub token has not been used on a request
- **THEN** its entry has no status

#### Scenario: Settings of the calling account

- **WHEN** an agent calls `get_settings` with a session or an API token
- **THEN** the result is that account's settings resolved against the defaults,
  with the defaults beside them

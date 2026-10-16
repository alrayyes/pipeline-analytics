# Spec Delta

## ADDED Requirements

### Requirement: Tools list, discover, register and untrack repositories

The system SHALL expose as WebMCP tools listing the tracked repositories,
listing a forge's tracked identifiers, discovering the repositories a token can
reach, registering a repository for tracking and untracking one, each sending
the same request as the Repositories page. Registering and untracking SHALL act
only after the person confirms.

#### Scenario: Listing repositories

- **WHEN** an in-browser agent calls `list_repos` without a limit
- **THEN** the first 20 repositories are requested, with `hasMore` in the result

#### Scenario: Discovering with the saved token

- **WHEN** an agent calls `discover_repos` with a forge and no token
- **THEN** the request carries no token, so the server uses the one saved for
  that forge, and nothing is registered

#### Scenario: Registering a repository

- **WHEN** an agent calls `register_repo` and the person confirms
- **THEN** the same POST as the page's Follow button is sent and the result is
  the repository the server returns, without a token

#### Scenario: The person declines to register or untrack

- **WHEN** an agent calls `register_repo` or `untrack_repo` and the person
  declines, or the page cannot ask
- **THEN** no request is sent and the call fails saying the person declined

#### Scenario: Untracking a repository

- **WHEN** an agent calls `untrack_repo` with an id from `list_repos` and the
  person confirms
- **THEN** the same DELETE as the page's untrack dialog is sent, with the id
  encoded once

#### Scenario: A missing argument

- **WHEN** an agent calls any of these tools without a required argument
- **THEN** the call fails saying which is required, nobody is asked and nothing
  is sent

#### Scenario: The server refuses

- **WHEN** the server answers with an error, such as 409 for a repository
  already tracked
- **THEN** the call fails and carries the status and the server's message

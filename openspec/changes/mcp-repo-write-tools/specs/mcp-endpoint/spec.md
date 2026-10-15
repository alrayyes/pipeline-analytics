# Spec Delta

## ADDED Requirements

### Requirement: Tools can register, discover and untrack repos

The system SHALL expose as MCP tools registering a repository, listing the
repositories a token can reach, and untracking a repository, each running the
same code as its endpoint and open to the same callers, an API token included.

#### Scenario: Registering a repo

- **WHEN** a client calls `register_repo` with a forge, an identifier and a
  token
- **THEN** the repo is tracked, the result carries its masked token and not the
  token, and the repo list endpoint shows it

#### Scenario: A repo that cannot be registered

- **WHEN** the repo is already tracked, or is archived, a fork or a mirror
- **THEN** the call reports `already_tracked`, `repo_archived`, `repo_fork` or
  `repo_mirror`, as the endpoint answers 409

#### Scenario: No token given

- **WHEN** a client calls `register_repo` or `discover_repos` without a token
  and none is saved for the forge and instance
- **THEN** the call reports `no_saved_token`; when one is saved, it is used

#### Scenario: Untracking

- **WHEN** a client calls `untrack_repo` for a tracked repo
- **THEN** the repo and its stored runs are removed, and an unknown id reports
  `not_found`

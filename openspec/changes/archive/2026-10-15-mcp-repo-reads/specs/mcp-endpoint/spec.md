# Spec Delta

## ADDED Requirements

### Requirement: Tools cover the service version and the tracked repos

The system SHALL expose as MCP tools the service version, the tracked
repositories with their filters, and a forge's tracked repository identifiers,
each calling the same code as the REST endpoint it matches and applying the
same validation.

#### Scenario: Listing repos never shows a token

- **WHEN** an agent calls `list_repos`
- **THEN** each repo carries its masked token only, as the endpoint returns it

#### Scenario: Identifiers need a known forge

- **WHEN** an agent calls `list_repo_identifiers` with no forge or an unknown
  one
- **THEN** the tool reports an error, as the endpoint answers 400

#### Scenario: Paging the repo list

- **WHEN** an agent calls `list_repos` with a limit smaller than the number of
  tracked repos
- **THEN** it gets that many repos and `hasMore` is true, and an offset
  continues from where the page ended

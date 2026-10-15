# mcp-endpoint Specification

## Purpose
Lets an agent do through MCP everything an API caller can do, apart from the operations the OpenAPI description marks `x-agent-excluded` with a reason (sign-in, tokens, passkeys, saved forge tokens, the webhooks and the probes). Reads return the same data as their endpoints; writes go through the same code, so the validation, the errors and the authorization are the endpoint's. A test fails when an operation has neither a tool nor a written exclusion.

## Requirements

### Requirement: MCP server reachable over Streamable HTTP

The system SHALL serve an MCP server at `POST /api/mcp` using the MCP
Streamable HTTP transport.

#### Scenario: A client connects to the MCP endpoint

- **WHEN** an MCP-capable client sends a Streamable HTTP request to
  `/api/mcp`
- **THEN** the system responds per the MCP Streamable HTTP transport
  and the client can list and call the tools this capability defines

### Requirement: MCP endpoint requires the same authentication as the REST API

The system SHALL require a valid session or a valid API token for every
`/api/mcp` request, identically to every other data endpoint.

#### Scenario: Unauthenticated MCP request is rejected

- **WHEN** a request to `/api/mcp` carries no valid session cookie and no
  valid bearer token
- **THEN** the system rejects it the same way it rejects an
  unauthenticated request to any other data endpoint

#### Scenario: Request bearing a valid API token succeeds

- **WHEN** a request to `/api/mcp` carries a valid, unrevoked API token as
  a bearer token
- **THEN** the system serves the request

### Requirement: MCP tools cover the existing read surface

The system SHALL expose, as MCP tools, at least: listing tracked
pipelines, fetching a single pipeline's health status and
duration/failure-rate trend, listing a pipeline's steps, listing a
pipeline's flaky runs, per-repository usage, and unhealthy steps across
all tracked pipelines.

#### Scenario: Listing tracked pipelines

- **WHEN** a client calls the tool that lists tracked pipelines
- **THEN** the result includes each pipeline's ID, name, and current
  health status, matching what `GET /api/pipelines` returns

#### Scenario: Fetching a pipeline's trend data

- **WHEN** a client calls the tool that fetches a single pipeline's
  detail
- **THEN** the result includes its duration trend and failure-rate
  trend, matching what `GET /api/pipelines/{pipelineId}` returns

### Requirement: Tool schemas stay in lockstep with the OpenAPI spec

The system SHALL derive each MCP tool's input schema and description
from `openapi/openapi.yaml`'s definition of the REST endpoint it
mirrors, rather than from a separately maintained schema.

#### Scenario: A REST endpoint's parameter changes

- **WHEN** a mirrored REST endpoint's request parameters change in
  `openapi/openapi.yaml`
- **THEN** the corresponding MCP tool's input schema reflects that
  change without a second, independent edit

### Requirement: Every operation has a tool or a written exclusion

The system SHALL fail its tests when an operation in the OpenAPI description,
of any method, has neither a tool nor an extension named `x-agent-excluded`
that states why an agent must not reach it.

#### Scenario: A write operation without a tool

- **WHEN** a POST, PUT, PATCH or DELETE operation is added to the OpenAPI
  description with no tool and no `x-agent-excluded` reason
- **THEN** the test suite fails and names the operation

#### Scenario: An exclusion with no reason

- **WHEN** an operation carries an empty `x-agent-excluded`
- **THEN** the test suite fails and names the operation

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

### Requirement: A write tool keeps its endpoint's authorization

The system SHALL let an MCP tool that writes act for exactly the callers its
endpoint admits. Where the endpoint is session-only (re-running or cancelling a
run, and quarantining a step), the tool SHALL act only when the call was
authenticated by a session and SHALL refuse a call authenticated by an API token
before anything is written, with the error the endpoint gives. Where the
endpoint takes an API token (registering, discovering and untracking repos, and
updating settings), so does the tool.

#### Scenario: Re-running a failed run with a session

- **WHEN** a client with a session calls `rerun_run` for a run that failed
- **THEN** the forge is asked to re-run the failed jobs and the result reports
  the request as accepted

#### Scenario: An API token on the run actions

- **WHEN** a client authenticated only by an API token calls `rerun_run` or
  `cancel_run`
- **THEN** the call reports `unauthenticated` and the forge is not asked

#### Scenario: A run that cannot be acted on

- **WHEN** a client with a session calls `rerun_run` for a run that has not
  concluded, or `cancel_run` for one that has
- **THEN** the call reports `not_actionable`, as the endpoint answers 409, and
  the forge is not asked

### Requirement: Tools can quarantine and un-quarantine a flaky step

The system SHALL expose as MCP tools marking a step as known for 30 days, with
an optional note of at most 500 characters, and clearing that mark, each acting
only for a signed-in session and reporting a failure as the endpoint does.

#### Scenario: Marking a step

- **WHEN** a client with a session calls `quarantine_step` for a flaky step
- **THEN** the step is quarantined for 30 days and the flaky list reports it

#### Scenario: A note that is too long

- **WHEN** a client with a session calls `quarantine_step` with a note over 500
  characters
- **THEN** the call reports `note_too_long` and nothing is marked

#### Scenario: An API token on the quarantine tools

- **WHEN** a client authenticated only by an API token calls `quarantine_step`
  or `unquarantine_step`
- **THEN** the call reports `unauthenticated` and the mark is unchanged

#### Scenario: Clearing a mark

- **WHEN** a client with a session calls `unquarantine_step` for a quarantined
  step, or for one with no mark
- **THEN** the step is not quarantined afterward and the call succeeds

#### Scenario: An unknown pipeline or an empty step

- **WHEN** a client with a session calls either tool for a pipeline that does
  not exist
- **THEN** the call reports `not_found`; with an empty step or pipeline ID it
  reports `invalid_input`, since the endpoint's path cannot match one

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
- **THEN** the repo and its stored runs are removed, and an unknown ID reports
  `not_found`

### Requirement: A tool can update the account's settings

The system SHALL expose as an MCP tool the partial update of the signed-in
account's settings, with the endpoint's meaning: a setting given a value is set,
one given null is cleared to its default, and one left out is unchanged.

#### Scenario: Changing one setting

- **WHEN** a client calls `update_settings` with only a theme
- **THEN** the theme changes, the other settings keep their values, and the
  result is the settings after the change

#### Scenario: Clearing a setting

- **WHEN** a client sets a setting to null
- **THEN** it returns to its documented default

#### Scenario: A value or key the settings do not allow

- **WHEN** a client gives a value outside a setting's set, or a key that is not
  a setting
- **THEN** the call is refused and nothing is stored

#### Scenario: An API token on the settings update

- **WHEN** a client authenticated by an API token calls `update_settings`
- **THEN** the call succeeds for that token's account, as the endpoint does

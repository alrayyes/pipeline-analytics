# forge-ingestion Specification

## Purpose
Collects GitHub Actions and Forgejo Actions workflow run, job, and step data for repositories the user chooses to track, via webhook events as the primary path and periodic reconciliation polling as a fallback, so the dashboard has accurate historical data at minimal API cost.

## Requirements

### Requirement: Repo tracking registration
The system SHALL let the user register a GitHub or Forgejo repository for tracking by supplying a repo-scoped personal access token (PAT) for that forge.

#### Scenario: Register a GitHub repo
- **WHEN** the user submits a GitHub repository identifier and a GitHub PAT scoped to that repository
- **THEN** the system stores the repository as tracked and associates it with the supplied PAT

#### Scenario: Register a Forgejo repo
- **WHEN** the user submits a Forgejo instance URL, repository identifier, and a Forgejo access token scoped to that repository
- **THEN** the system stores the repository as tracked and associates it with the supplied token

### Requirement: Credential storage
The system SHALL store every GitHub and Forgejo access token encrypted at rest and SHALL NOT display a stored token in full after initial submission.

#### Scenario: Token not recoverable in plaintext
- **WHEN** the user views a tracked repo's settings after registration
- **THEN** the system shows only a masked/truncated representation of the stored token, never the full value

### Requirement: Webhook registration on tracking
The system SHALL create a webhook on a repository when it is registered for tracking, subscribed to workflow run and job events, secured with a per-repo signing secret.

#### Scenario: Webhook created on registration
- **WHEN** a repository is successfully registered for tracking
- **THEN** the system creates a webhook on that repository via the forge's API, subscribed to workflow run/job events, with a unique signing secret

#### Scenario: Webhook creation failure is surfaced
- **WHEN** webhook creation fails (for example, insufficient token scope)
- **THEN** the system marks the repository's ingestion status as degraded and reports the failure reason to the user

### Requirement: Webhook event ingestion
The system SHALL verify the signature of every incoming webhook delivery against the repository's signing secret and SHALL reject unverified deliveries.

#### Scenario: Valid webhook delivery is ingested
- **WHEN** a workflow run or job event is delivered with a valid signature for a tracked repository
- **THEN** the system creates or updates the corresponding run, job, and step records with the event's data

#### Scenario: Invalid signature is rejected
- **WHEN** a webhook delivery's signature does not match the repository's signing secret
- **THEN** the system discards the payload and does not modify any stored run data

### Requirement: Reconciliation polling

The system SHALL poll each tracked repository's recent workflow runs
at least hourly, using conditional requests (ETag/If-None-Match on
GitHub) to avoid rate-limit cost when nothing has changed, to backfill
history and catch missed webhook deliveries.

#### Scenario: Reconciliation backfills a newly tracked repo

- **WHEN** a repository is registered for tracking
- **THEN** the next reconciliation poll fetches and stores its recent
  workflow run history, not only future webhook events

#### Scenario: Unchanged repo costs no rate-limit budget

- **WHEN** a reconciliation poll runs against a GitHub repository with
  no new activity since the last poll
- **THEN** the system issues a conditional request and, on receiving a
  304 response, does not modify stored data and does not count the
  request against the token's primary rate limit

#### Scenario: Reconciliation catches a missed webhook

- **WHEN** a workflow run completed on the forge but no corresponding
  webhook delivery was ingested
- **THEN** the next reconciliation poll detects the discrepancy and
  stores the missing run, job, and step data

#### Scenario: An already-completed run's jobs aren't refetched

- **WHEN** a reconciliation poll's run listing reports a run as
  `completed` with the same conclusion already stored for it
- **THEN** the system does not issue a request for that run's jobs,
  even if other runs in the same poll did change

### Requirement: Duplicate registration is rejected clearly

The system SHALL reject an attempt to register a repository identifier
that is already tracked on the same forge (and, for Forgejo, the same
instance URL) with a distinct response indicating it's already
tracked, rather than a generic internal error.

#### Scenario: Re-registering an already-tracked repo

- **WHEN** the user submits a registration for a repository identifier
  already tracked on that forge and instance
- **THEN** the system rejects it with a response identifying it as
  already tracked, and does not create a second record for it

#### Scenario: The registration UI never offers an already-tracked repo

- **WHEN** the user runs repository discovery
- **THEN** the system excludes every already-tracked repository on
  that forge and instance from the results, regardless of how many
  repositories are tracked in total

### Requirement: A malformed Forgejo response doesn't crash ingestion

The system SHALL skip a null run, job or step in a Forgejo response and keep
the surrounding entries, and SHALL NOT panic on any response body.

#### Scenario: A null run

- **WHEN** the run list response contains `null` in place of a run
- **THEN** the poll succeeds and returns the other runs

#### Scenario: A null job or step

- **WHEN** a job list response contains `null` in place of a job, or the steps
  of a job contain `null`
- **THEN** that entry is skipped and the rest are kept

### Requirement: Decoders of external data are fuzzed

The system SHALL keep a fuzz test seeded with real payloads for each decoder of
external data, replay committed crashing inputs in every `go test` run, and
run each fuzz test for a bounded time on a schedule.

#### Scenario: A crashing input is found

- **WHEN** a scheduled fuzz run finds an input that panics a decoder
- **THEN** the job fails and uploads the input for committing under
  `testdata/fuzz`

#### Scenario: A committed crasher

- **WHEN** a crashing input is committed under `testdata/fuzz`
- **THEN** an ordinary `go test` replays it

### Requirement: A forge token can be saved and reused

The system SHALL let a signed-in session save one forge token per forge and
Forgejo instance, store it encrypted, return only its masked form, and use it
for registration and discovery when the request carries no token.

#### Scenario: Registering with the saved token

- **WHEN** a repo is registered with no token and a token is saved for its
  forge and instance
- **THEN** the saved token is used, and no response contains it

#### Scenario: No saved token

- **WHEN** a request carries no token and none is saved
- **THEN** the answer is `400` with code `no_saved_token`

#### Scenario: Replacing and deleting

- **WHEN** a token is saved for a forge and instance that already has one
- **THEN** it replaces it, and deleting it leaves already registered repos
  working

#### Scenario: An API token

- **WHEN** an API token, not a session, calls the saved token endpoints
- **THEN** the answer is `401`

### Requirement: Run commit metadata

The system SHALL record, for each ingested run, its branch, head
commit SHA, commit message first line and triggering actor when the
forge provides them, from both webhook events and reconciliation
polling.

#### Scenario: Webhook run carries commit metadata

- **WHEN** a run webhook arrives with head commit and actor fields
- **THEN** the stored run has its branch, SHA, message and actor

#### Scenario: Forge omits a field

- **WHEN** the forge payload lacks the actor
- **THEN** the run is stored with no actor and ingestion does not
  fail

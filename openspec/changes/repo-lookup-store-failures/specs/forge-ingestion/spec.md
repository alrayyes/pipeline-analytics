# Spec Delta

## ADDED Requirements

### Requirement: A store failure is not reported as a missing repository

The system SHALL answer 404, or a not-found tool error, for a repository
lookup or removal only when the store reports `ErrRepoNotFound`. Any other
store error SHALL be logged and answered 500 with the `internal_error` body,
or returned as a tool error, without the error's detail. This holds for the
repo usage and untrack endpoints and tools, the job log and the run actions.

#### Scenario: The store fails while reading usage

- **WHEN** the store returns an error other than `ErrRepoNotFound` for a
  usage request
- **THEN** the response is 500 with the `internal_error` error body

#### Scenario: The store fails while untracking

- **WHEN** the store returns an error other than `ErrRepoNotFound` on removal
- **THEN** the response is 500 and the MCP tool reports `internal_error`

#### Scenario: A job's repo is gone

- **WHEN** a job log or run action finds its repo untracked
- **THEN** the job or run is reported not found

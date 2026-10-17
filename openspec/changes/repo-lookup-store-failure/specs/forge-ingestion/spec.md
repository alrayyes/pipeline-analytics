# Spec Delta

## ADDED Requirements

### Requirement: A repository lookup tells a missing repository from a store failure

The system SHALL answer 404 to `GET /api/repos/{repoId}` only when the store
reports no repository under the id, through the `Store` port's not-found
sentinel. For any other store error it SHALL log the error and answer 500 with
the API's `internal_error` body, without the error's detail.

#### Scenario: A store failure

- **WHEN** a caller requests a repository and the store fails with an error
  that isn't the not-found sentinel
- **THEN** the response is 500 with the `internal_error` body and the error is
  logged

#### Scenario: The store's not-found return

- **WHEN** the store is asked for an id no repository is tracked under
- **THEN** it returns the port's not-found sentinel

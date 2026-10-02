# Spec Delta

## ADDED Requirements

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

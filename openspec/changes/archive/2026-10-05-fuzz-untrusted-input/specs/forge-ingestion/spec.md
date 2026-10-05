# Spec Delta

## ADDED Requirements

### Requirement: Decoders of external data are fuzzed

The system SHALL keep a fuzz test seeded with real payloads for each decoder of
external data, replay committed crashing inputs in every `go test` run, and
run the fuzzers for a bounded time on a schedule.

#### Scenario: A crashing input is found

- **WHEN** a scheduled fuzz run finds an input that panics a decoder
- **THEN** the job fails and uploads the input for committing under
  `testdata/fuzz`

#### Scenario: A committed crasher

- **WHEN** a crashing input is committed under `testdata/fuzz`
- **THEN** an ordinary `go test` replays it

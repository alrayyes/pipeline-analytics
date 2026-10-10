# Proposal

## Why

#471 asked what "Quarantine" should mean. The telemetry design shows the
button beside Re-run and Cancel, but nobody said what it does. The decision
is to build a small version: a quarantined step keeps showing, but a known
flake stops turning its pipeline unhealthy.

## What Changes

- A signed-in user can **quarantine** a flaky step, a step being one name
  within one pipeline, and **un-quarantine** it. The mark lives in this app
  only: no forge is read or written for it.
- A quarantined step **stays in the flaky list**, labelled "quarantined", with
  the age of the mark and an optional note.
- A quarantined step **stops raising the `flaky_step` health signal** and
  **stops counting in `FlakyStepRatio`**. Nothing else changes: the step's
  failure rate, run counts, result matrix, and the pass and failure rates of
  its runs are computed as before. A quarantined step that also fails
  consistently still raises the failure-rate signal through its runs.
- A quarantine **expires after 30 days** unless renewed, so a mark can't be
  forgotten. An expired mark behaves as if it was never set.
- The REST API and the MCP read tools report a `quarantined` field on flaky
  steps. Setting and clearing is **session-only**: an API token or an MCP
  client can't quarantine anything, the same position as saved tokens and the
  forge actions.

## Rejected

- **Hide the step from the flaky list and insights.** A quarantined step is
  forgotten, and a real regression in it can't be seen.
- **Exclude the step from pass rate.** The failure rate is the one figure the
  whole app agrees on; muting it makes every number an asterisk.
- **Tag only.** It silences nothing, so nobody has a reason to use it.

## Capabilities

### New Capabilities

- `step-quarantine`: marking a flaky step quarantined and un-quarantined, what
  a mark does and doesn't change, its 30-day expiry, and who may write it.

### Modified Capabilities

- `dashboard-ui`: the flaky-step callout and the flaky test telemetry show the
  quarantined label, and a quarantined step no longer makes its pipeline
  unhealthy.

## Impact

- `internal/metrics`: `computeHealth`'s `hasFlakyStep` and `flakyStepRatio`
  skip quarantined steps; `ListFlakySteps` carries the mark.
- `internal/db`: a migration for a quarantine table keyed on pipeline and step
  name. Health stays recomputed on every request (ADR 0004); the table is
  joined at read time, so no health state is stored.
- `internal/httpserver`: a session-only write route, a `quarantined` field on
  the flaky-step responses, `openapi/openapi.yaml`, and the MCP flaky tools'
  output (the spec-parity test pins both).
- `web/`: the list and detail UI. A frontend ticket of its own.
- Needs the Stitch telemetry designs in `docs/design/` read for the button
  and label placement.

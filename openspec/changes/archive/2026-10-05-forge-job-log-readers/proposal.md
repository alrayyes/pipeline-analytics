# Proposal

## Why

The log endpoint specified in #449 needs the forge side first: something that
fetches a job's log tail from GitHub, and says plainly that Forgejo can't.
Tracked in #343. Independent of #449 and #450.

## What Changes

- `ingestion.JobLogReader`, separate from `ForgeClient` so no existing
  implementation or fake has to change, and `ErrLog*` errors for the four
  ways a log can be missing.
- GitHub: follows the log redirect server-side, streams the body and keeps the
  last N lines, each cut to 4096 bytes. Reading stops at 64 MiB.
- Forgejo: reports `ErrLogUnsupported`, since the API landed in v16 and this
  service targets 11.0.16.

## Capabilities

### Modified Capabilities

- `job-logs`: how a forge supplies a job's log tail.

# Proposal

## Why

The log endpoint specified in #449 is addressed by job, and a step in
`GET /api/runs/{runId}/steps` didn't say which job it ran in. Tracked in #343.
Stacked on #449.

## What Changes

- Run steps carry `jobId`, in the run detail and the run list.

## Capabilities

### Modified Capabilities

- `job-logs`: a step names the job whose log to ask for.

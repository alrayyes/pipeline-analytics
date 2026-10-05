# Proposal

## Why

The log endpoint specified in #449 has its `jobId` (#450) and its forge
readers (#451) but no handler. This wires them together. Tracked in #343,
backend half.

## What Changes

- `GET /api/runs/{runId}/jobs/{jobId}/log` is served. Without a log service in
  its deps the route isn't registered, so a build without one answers 404.
- `ingestion.JobLogService` resolves the job, picks the repo's forge reader
  and returns the tail, or a reason. Unsupported when a forge has no reader.
- The store resolves a run and job id pair to the job's forge id, link and
  repo. A job under another run is not found.
- `lines` outside 1 to 1000 is a `400`.
- The MCP tool for it is the next change; the parity test lists the route
  until then.

## Capabilities

### Modified Capabilities

- `job-logs`: the endpoint behaves as the spec says.

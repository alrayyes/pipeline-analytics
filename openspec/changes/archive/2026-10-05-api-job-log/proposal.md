# Proposal

## Why

The dashboard links to the forge for a failing step's log, which costs a
phone user a context switch. Tracked in #343, backend half. The frontend
session builds the UI against the contract below, so the contract goes up
first and is reviewed as the design.

## Research

- GitHub: `GET /repos/{owner}/{repo}/actions/jobs/{job_id}/logs` answers 302
  to a plain-text download that expires after a minute
  ([docs](https://docs.github.com/en/rest/actions/workflow-jobs)). It is
  per job. GitHub has no per-step log endpoint.
- Forgejo: the Actions log REST API landed in v16
  ([forgejo/forgejo#12666](https://codeberg.org/Cyborus/forgejo-api/pulls/155)).
  Earlier versions serve logs only through the web UI. The instance this
  service targets runs 11.0.16, so it has none.

## Decisions

- **Fetch on demand, never store.** Storing grows the SQLite file with log
  volume and outlives the forge's retention. Fetching needs the forge token at
  request time, which the service already holds, encrypted, per repo.
- **Per job, not per step**, because neither forge exposes a step's log on
  its own. The step detail passes its job's id.
- **Unavailable is a `200`, not an error**, so the client says why and still
  shows the deep link (criterion 3).
- **ANSI passes through.** The server doesn't render or strip it; the client
  renders colours and escapes everything else (criterion 2).

## What Changes

- New `GET /api/runs/{runId}/jobs/{jobId}/log` and a `JobLog` schema.
- `RunStep` gains `jobId`.
- No handler yet: it lands in the next change, which also adds the MCP tool.

## Capabilities

### New Capabilities

- `job-logs`: reading a job's log tail through the API.

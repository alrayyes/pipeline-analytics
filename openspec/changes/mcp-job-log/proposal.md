# Proposal

## Why

`mcp_spec_parity_test.go` requires every read endpoint to have an MCP tool or
a stated reason. The job log endpoint (#455) is useful to an agent triaging a
failure, so it gets a tool. Tracked in #343, backend half. Stacked on #455.

## What Changes

- `get_job_log(runId, jobId, lines?)` returns the same body as the REST
  endpoint. A bad `lines` or an unknown job is a tool error.
- A build with no log service registers the tool and answers with an error,
  so the tool list is the same everywhere.
- README and `llms.txt` count the tools correctly (the latter said six).

## Capabilities

### Modified Capabilities

- `job-logs`: the log is readable through MCP.

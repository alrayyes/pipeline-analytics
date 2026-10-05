# Spec Delta

## ADDED Requirements

### Requirement: The job log is readable through MCP

The system SHALL expose a read-only `get_job_log` tool returning what
`GET /api/runs/{runId}/jobs/{jobId}/log` returns, and SHALL report an unknown
job or an out-of-range `lines` as a tool error.

#### Scenario: A job's log

- **WHEN** a client calls `get_job_log` with a run and job id
- **THEN** the result carries `available`, `lines`, `truncated` and `forgeUrl`

#### Scenario: An unknown job

- **WHEN** the job doesn't exist
- **THEN** the call is a tool error

# Spec Delta

## Purpose

Exposes the pipeline health, duration/failure-rate trend, flaky-step,
and usage data the dashboard and REST API already serve to MCP-capable
agent clients, over a standard MCP transport, kept in lockstep with the
REST API it mirrors instead of drifting as its own hand-maintained tool
list.

## ADDED Requirements

### Requirement: MCP server reachable over Streamable HTTP

The system SHALL serve an MCP server at `POST /mcp` using the MCP
Streamable HTTP transport.

#### Scenario: A client connects to the MCP endpoint

- **WHEN** an MCP-capable client sends a Streamable HTTP request to
  `/mcp`
- **THEN** the system responds per the MCP Streamable HTTP transport
  and the client can list and call the tools this capability defines

### Requirement: MCP endpoint requires the same authentication as the REST API

The system SHALL require a valid session or a valid API token for every
`/mcp` request, identically to every other data endpoint.

#### Scenario: Unauthenticated MCP request is rejected

- **WHEN** a request to `/mcp` carries no valid session cookie and no
  valid bearer token
- **THEN** the system rejects it the same way it rejects an
  unauthenticated request to any other data endpoint

#### Scenario: Request bearing a valid API token succeeds

- **WHEN** a request to `/mcp` carries a valid, unrevoked API token as
  a bearer token
- **THEN** the system serves the request

### Requirement: MCP tools cover the existing read surface

The system SHALL expose, as MCP tools, at least: listing tracked
pipelines, fetching a single pipeline's health status and
duration/failure-rate trend, listing a pipeline's steps, listing a
pipeline's flaky runs, per-repository usage, and unhealthy steps across
all tracked pipelines.

#### Scenario: Listing tracked pipelines

- **WHEN** a client calls the tool that lists tracked pipelines
- **THEN** the result includes each pipeline's ID, name, and current
  health status, matching what `GET /api/pipelines` returns

#### Scenario: Fetching a pipeline's trend data

- **WHEN** a client calls the tool that fetches a single pipeline's
  detail
- **THEN** the result includes its duration trend and failure-rate
  trend, matching what `GET /api/pipelines/{pipelineId}` returns

### Requirement: No write tool

The system SHALL NOT expose any MCP tool that creates, modifies, or
deletes data.

#### Scenario: Tool list contains no mutating tool

- **WHEN** a client lists the tools this capability exposes
- **THEN** none of them registers a repository, untracks a repository,
  or otherwise changes stored state

### Requirement: Tool schemas stay in lockstep with the OpenAPI spec

The system SHALL derive each MCP tool's input schema and description
from `openapi/openapi.yaml`'s definition of the REST endpoint it
mirrors, rather than from a separately maintained schema.

#### Scenario: A REST endpoint's parameter changes

- **WHEN** a mirrored REST endpoint's request parameters change in
  `openapi/openapi.yaml`
- **THEN** the corresponding MCP tool's input schema reflects that
  change without a second, independent edit

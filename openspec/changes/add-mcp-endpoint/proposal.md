# Proposal

## Why

pipeline-analytics' health/duration/failure-rate/flaky-step/usage data is
reachable only through its REST API (read by a human via the dashboard, or
by the generated SDKs) — nothing exposes it to an MCP-capable agent. An
agent doing CI/CD triage ("why did this repo's pipeline get slower this
week", "which step is flaky enough to fix first") has to have a human read
the dashboard and relay the numbers by hand. Tracked as
[alrayyes/pipeline-analytics#295](https://github.com/alrayyes/pipeline-analytics/issues/295).

## What Changes

- New `POST /api/mcp` endpoint (bearer-token-gated, same as every other data
  API) serving an MCP server over the Streamable HTTP transport, built on
  the official `github.com/modelcontextprotocol/go-sdk` — no hand-rolled
  JSON-RPC framing.
- Read-only MCP tools covering the same ground the dashboard and REST API
  already do: list tracked pipelines, get a pipeline's health status and
  duration/failure-rate trends, list its steps (with flaky/slow signals),
  list its flaky runs, per-repo usage (runner-minutes), and unhealthy
  steps across all tracked pipelines. No write tool — the only write path
  in the existing API is repo registration/untracking, out of scope per
  the issue.
- Each tool's input schema and description are generated from
  `openapi/openapi.yaml`'s existing path/parameter definitions for the
  equivalent REST endpoint, so the tool list can't drift from the spec it
  mirrors — not a second, hand-maintained list.
- `openapi/openapi.yaml` gains the `/api/mcp` path itself (method, security
  requirement, request/response envelope), documented like every other
  endpoint, even though its body is MCP's own JSON-RPC framing rather
  than a REST payload — consistent with "every HTTP endpoint is in the
  spec," and it's also the source generation reads from.
- README gets an "MCP" section next to the existing API/installation
  docs: the endpoint URL, that it needs the same bearer token as the
  REST API, and how to point an MCP-capable client at it.

## Capabilities

### New Capabilities

- `mcp-endpoint`: an MCP server, reachable over Streamable HTTP at
  `/api/mcp`, exposing the system's existing pipeline-health/metrics data as
  read-only MCP tools kept in lockstep with `openapi/openapi.yaml`.

### Modified Capabilities

(none — `dashboard-auth`'s existing requirement that every API endpoint
accept a valid session or API token already covers `/api/mcp` without a
wording change, and `http-observability`'s per-request logging
requirement already covers it the same way every other endpoint is
covered.)

## Impact

- `internal/httpserver/server.go` (mount the MCP handler on the mux,
  inside the same `requireSession` gate every other data endpoint sits
  behind).
- New `internal/httpserver/mcp.go` (or a dedicated `internal/mcp`
  package, decided in design.md) wiring MCP tools to `deps.Metrics` /
  `deps.IngestionStore`, the same service interfaces the REST handlers
  already call — no new domain logic, no new port.
- `go.mod` / `go.sum` (new pinned dependency on
  `github.com/modelcontextprotocol/go-sdk`).
- `openapi/openapi.yaml` (new `/api/mcp` path).
- `README.md` (new MCP section).
- No SDK repo changes — the generated `pipeline-analytics-sdk-*` clients
  are REST clients; an MCP client is a different kind of consumer
  entirely and isn't what those repos generate.

# Design

## Context

See proposal.md for motivation. Relevant current state:

- `internal/httpserver` is a flat set of handler files (`pipelines.go`,
  `repos.go`, `insights.go`, …), each converting a domain type from
  `internal/metrics` / `internal/ingestion` into a JSON DTO and wired
  onto the shared `*http.ServeMux` in `server.go`. No domain logic lives
  in this package — it's the adapter layer the ARCHITECTURE.md hexagonal
  shape describes.
- `requireSession` already gates every non-public path behind a session
  cookie or bearer API token (`internal/httpserver/session.go`), and
  `requestLogger` already logs every request (`internal/httpserver/logging.go`).
  Both wrap the mux as a whole, so anything mounted on the mux inherits
  both for free.
- `deps.Metrics *metrics.Service` and `deps.IngestionStore
  ingestion.Store` are the two service objects the existing read
  endpoints call directly (`pipelines.go`, `repos.go`'s usage handler,
  `insights.go`). No new domain method is needed — every data point the
  proposal's six tools need is already exposed by one of these two.

## Goals / Non-Goals

**Goals:**

- Serve the six read-only tools listed in `specs/mcp-endpoint/spec.md`
  over MCP Streamable HTTP at `/mcp`, behind the same auth as the REST
  API.
- Keep each tool's input schema demonstrably in sync with
  `openapi/openapi.yaml`'s definition of the endpoint it mirrors.

**Non-Goals:**

- No write/mutating tool (proposal and spec already rule this out).
- No new domain capability — every tool is a thin translation of an
  existing `metrics.Service` / `ingestion.Store` call, the same
  translation `internal/httpserver`'s REST handlers already do.
- No support for MCP transports other than Streamable HTTP (stdio, SSE)
  — this is a server reachable over the network like the rest of the
  API, not a local process a client spawns.

## Decisions

### Use the official `github.com/modelcontextprotocol/go-sdk` for the MCP layer itself

Per this account's "check for an existing SDK before hand-rolling an
integration": don't hand-write MCP's JSON-RPC framing, session
handling, or the Streamable HTTP transport. `go-sdk`'s `mcp` package
provides a server that mounts as a plain `http.Handler`
(`mcp.NewStreamableHTTPHandler`), which drops straight onto the
existing mux next to every other handler in `server.go`.

**Alternative considered**: a runtime OpenAPI-to-MCP bridge library
(e.g. `alexliesenfeld/openapimcp`, `ubermorgenland/openapi-mcp`) that
reads `openapi.yaml` directly and serves matching MCP tools with no
Go code per tool — attractive because it would make "kept in lockstep
with the spec" automatic rather than test-enforced. Rejected for now:
every such library found has no tagged release (pseudo-versions off a
commit hash) and no evident track record, which is a heavier
dependency to run behind a public-facing, authenticated endpoint than
this project's other dependencies. Worth revisiting once one of these
matures — noted as an open question below, not a blocker.

### Hand-write the six tools against the existing service objects, not a generator

Each tool is a small function in a new `internal/httpserver/mcp.go`,
following the same shape as `pipelines.go`'s handlers: call
`deps.Metrics` / `deps.IngestionStore`, convert the result to the
tool's output type, return it. No HTTP round-trip to the REST API
itself (no self-loopback call) — same in-process call the REST handler
for the equivalent endpoint already makes.

**Alternative considered**: generate tool registration code from
`openapi.yaml` at build time (a `go:generate` step). Rejected as
disproportionate for six tools — the generator would be more code than
what it generates, and it would need updating for every new tool
shape anyway.

### Enforce "kept in lockstep with the spec" with a parity test, not generation

A new test (`internal/httpserver/mcp_spec_parity_test.go`) parses
`openapi/openapi.yaml` with `github.com/getkin/kin-openapi` (new
dev-only-in-spirit but regular dependency; well-maintained, the same
category of "official-enough" OpenAPI tooling this project already
leans on for `redocly lint`) and asserts, for each of the six tools,
that its input schema's parameter names and required/optional status
match the corresponding path's declared parameters. The test fails
if a REST parameter is added, renamed, or has its required-ness
changed without the matching MCP tool schema being updated in the same
commit — the proposal's "generated from (or kept in lockstep with)"
satisfied by a failing test rather than by construction.

### Where the `/mcp` path lives in the spec

`openapi/openapi.yaml` gains a `/mcp` path entry whose request/response
bodies are typed loosely (MCP's JSON-RPC envelope, not a REST payload)
— documented for discoverability and consistency with "every endpoint
is in the spec" (`CONTRIBUTING.md`), not because Redocly-driven client
generation applies to it. The `pipeline-analytics-sdk-*` repos generate
REST clients from this spec; an MCP entry with a JSON-RPC body isn't
something those generators need to handle specially, but the SDK repos
aren't touched by this change regardless (proposal's Impact section).

### Auth: reuse `requireSession` as-is

`/mcp` is mounted on the same mux, inside the same `requireSession`
wrapper, as every other data endpoint — no new code in
`session.go`/`auth.go`. `dashboard-auth`'s existing requirement already
covers it; see proposal's "Modified Capabilities: (none)".

## Risks / Trade-offs

- **New dependency surface** (`go-sdk`, `kin-openapi`) → both are
  widely used (`go-sdk` is Google-co-maintained per the MCP org itself;
  `kin-openapi` is the de facto standard OpenAPI 3 Go parser, already
  the base of several tools this search turned up). Pinned to an exact
  version per this account's dependency rule.
- **Hand-written tool schemas can still drift between the parity test
  running and a reviewer actually reading the diff** → the test only
  catches parameter name/required-ness drift, not a semantic
  description drift (e.g. the REST doc changes what a field means but
  keeps its name). Accepted: matches the level of rigor
  `redocly lint` already gives the REST spec itself — structural, not
  semantic.
- **A large tracked-pipeline count returned to an MCP client in one
  tool call** → the REST endpoints these tools mirror already paginate
  (`pipelineListDTO.HasMore`); the tools surface the same pagination
  parameters rather than trying to return everything at once.

## Open Questions

- Whether to switch to a runtime OpenAPI-to-MCP bridge library once one
  reaches a tagged release, replacing the hand-written tools and the
  parity test with generation-by-construction. Doesn't change this
  change's specs, approach, or tasks — a future change's call.

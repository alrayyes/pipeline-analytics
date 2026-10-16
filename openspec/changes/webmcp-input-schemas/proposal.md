# Proposal

## Why

No WebMCP tool declares an input schema, so an agent has to guess
`pipelineId`, `runId`, `jobId` and the allowed values (#649).
`rules/mcp.md` wants each schema derived from the OpenAPI spec rather than
copied by hand, and `list_runs` pages by default where the API does not, which
its description never said.

## What Changes

- `web/scripts/generate-webmcp-schemas.ts` builds each tool's JSON Schema from
  the path and query parameters of the operation it wraps, and writes
  `webmcpSchemas.generated.ts`, which is committed.
- `webmcpOperations.ts` names the operation behind each tool, shared by the
  generator and the parity test.
- Every tool registers with its generated `inputSchema`.
- A unit test regenerates the module in memory and fails when the committed
  copy or a tool's schema is stale, so a changed spec parameter turns it red.
- `list_pipelines` and `list_runs` state their default paging (20 from offset
  0) in their description and schema. Every description now says where its ids
  come from.

## Capabilities

### Modified Capabilities

- `webmcp-support`: every tool carries an input schema generated from the spec.

## Impact

- `web/src/lib/webmcpTools.ts`, `webmcpOperations.ts`, the generated module,
  `web/scripts/generate-webmcp-schemas.ts` and their tests, `CONTRIBUTING.md`.

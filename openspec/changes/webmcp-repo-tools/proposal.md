# Proposal

## Why

Listing, discovering, registering and untracking repositories has no WebMCP
tool, though the MCP server has all of them (the parity work, `rules/mcp.md`).
An in-browser agent sharing the Repositories page cannot do what the page does.

## Decisions

- **The tools send the Repositories page's own requests.** The page built each
  request inline, so they move to `repoApi.ts` first and the page and the tools
  both call it. The endpoint applies its own validation and conflict codes. A
  refusal reaches the agent as the status and the server's own message, which
  says what to fix.
- **Registering and untracking ask the person first**, as re-run and cancel do.
  Registering stores a token and creates a webhook on the forge; untracking
  deletes the repository's stored runs, jobs and steps. Both are annotated
  `destructiveHint`. The question names the repository and what the call does,
  and never repeats a token.
- **Discovery and the two reads do not ask.** `discover_repos` registers
  nothing and stores nothing; the page runs it from a form without a dialog.
- **A token is an optional input, as it is on the MCP server.** Omitting it uses
  the token saved for the forge and instance, so an agent need not handle one.
  The result never contains it; the server does not echo it.
- **`repoId` is passed unencoded.** The tool encodes it. The generated
  description for it says so, as for a step name.
- **Paging matches the other list tools:** `list_repos` returns 20 from offset 0
  unless told otherwise, where the endpoint returns everything.

## What Changes

- `list_repos`, `list_repo_identifiers`, `discover_repos`, `register_repo` and
  `untrack_repo`.
- The Repositories page's requests move to `repoApi.ts`; the page behaves as
  before.
- `ApiError` moves to its own module so a test that replaces `dashboardApi.js`
  does not take it away.
- The five operations come off the parity test's pending list.

## Capabilities

### Modified Capabilities

- `webmcp-support`: the tools cover tracking repositories.

## Impact

- `web/src/lib/repoApi.ts`, `apiError.ts`, `webmcpTools.ts`,
  `webmcpOperations.ts`, the generated schemas, the Repositories page and their
  tests.

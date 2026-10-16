# Proposal

## Why

Reading and changing the account's dashboard settings has no WebMCP tool,
though the MCP server has `get_settings` and `update_settings` (the parity work,
`rules/mcp.md`). An in-browser agent cannot set the theme or the Pipelines
filters the way the page's own controls do.

## Decisions

- **The tools send the stores' own requests.** The settings requests were
  `fetchSettings` and `patchSettings`, which swallow every failure on purpose
  because a store has already updated itself. They now sit on `loadSettings` and
  `saveSettings`, which throw an `ApiError` with the server's message, and the
  two best-effort functions wrap them unchanged. The tools use the throwing pair,
  so an agent sees why a value was refused.
- **`ApiError` carries the server's code and message.** The repo requests kept
  these in a subclass; the settings requests need them too, so the base class
  has them and the subclass goes.
- **No confirmation prompt.** These are the signed-in user's own display
  preferences. The page saves each one on a click without asking, nothing
  leaves this service, and setting one back undoes it. This matches the
  quarantine tools, not re-run and untrack.
- **A change shows on the open page.** The stores seed from the layout's load, so
  a PATCH made behind their back would leave the page stale until the next
  navigation. The layout hands the registration `invalidateAll`, and the tool
  calls it after a successful update. A failed refresh does not fail the call:
  the change is saved.
- **The server validates.** The input schema is generated from the spec's
  `SettingsUpdate` (closed sets with `null`, the repo selector's length), and the
  tool forwards whatever it is given, so an unknown key or value is the
  endpoint's 400 and the whole update is refused, as on the endpoint.
- **A null is meaningful.** It clears a setting to its default, as the endpoint
  documents; each property's description says so.

## What Changes

- `get_settings` and `update_settings`.
- `registerWebMCPTools` takes an optional `refreshPage`; the root layout passes
  `invalidateAll`.
- `loadSettings` and `saveSettings` in `settingsSync.ts`; `ApiError` carries the
  server's code and message.
- Both operations come off the parity test's pending list.

## Capabilities

### Modified Capabilities

- `webmcp-support`: the tools cover reading and changing settings.

## Impact

- `web/src/lib/settingsSync.ts`, `apiError.ts`, `repoApi.ts`, `webmcpTools.ts`,
  `webmcpOperations.ts`, the generated schemas, the root layout and their tests.

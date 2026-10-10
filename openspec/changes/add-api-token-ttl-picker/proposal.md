# Proposal

## Why

`POST /api/auth/tokens` issues every API token with a fixed 365-day
lifetime (`apiTokenTTL` in `internal/auth/sqlite/store.go`), a deliberate
choice when token auth shipped (#178) but too blunt: a short-lived CI
credential shouldn't carry a year-long blast radius. Tracked as
[#191](https://github.com/alrayyes/pipeline-analytics/issues/191).

There is no token UI at all. Tokens are created with `curl` (see the
README), so the picker also needs somewhere to live.

This change was first written while token auth was unmerged, and #191 was
closed by its proposal pull request before anything was built. It is
revised here against what exists now: token auth is merged, and a Settings
page exists.

## What Changes

- `POST /api/auth/tokens` accepts an optional `ttlSeconds`. The server
  clamps it to a 365-day ceiling, rejects zero or less with `400`, and uses
  a 90-day default when it is omitted. The response carries `expiresAt`.
  The ceiling is the enforcement boundary; the client's presets are UX only.
- **The default drops from 365 to 90 days** for a caller that sends no
  body. A script that relied on a year must now ask for `ttlSeconds`.
- An "API tokens" section on the Settings page: presets of 30 days, 90 days
  and 1 year (90 selected), a live expiry preview, and the token shown once
  on creation. It does not list or revoke tokens.
- No "no expiration" option anywhere.
- `openapi/openapi.yaml` gains the optional request field and the response's
  `expiresAt`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-auth`: "API token issuance and revocation" gains the
  client-selectable, server-clamped lifetime, the default and the `400`.
- `dashboard-ui`: a new requirement, the API tokens section on Settings.

## Impact

- `internal/auth`: `CreateToken` and `IssueToken` take a requested
  lifetime; `apiTokenTTL` becomes `apiTokenTTLCeiling` and
  `apiTokenTTLDefault` is added.
- `internal/httpserver/auth.go`: the handler reads `ttlSeconds` and returns
  `expiresAt`.
- `openapi/openapi.yaml`, the README's token instructions.
- Frontend, a separate ticket: the Settings section and its tests. It is not
  in this change's backend pull request.
- Accessibility: the picker needs the `role="radiogroup"` wrapper that
  `ForgeFilter.svelte` uses and an axe-core scan.

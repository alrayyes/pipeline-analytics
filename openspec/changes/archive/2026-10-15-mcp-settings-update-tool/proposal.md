# Proposal

## Why

Changing the account's settings has no MCP tool (#615), the last operation the
Go MCP server lacks. `rules/mcp.md` wants every operation reachable through
MCP, with the same authorisation and validation as the endpoint.

## Decisions

- **The endpoint's body, including null.** A setting present with a value is
  set, present with `null` is cleared to its default, and left out is
  untouched. A struct cannot tell null from absent, so the tool's argument is a
  map, passed straight to `settings.Service.Update`, which already validates
  keys and values.
- **An explicit input schema, built from the domain.** Each property is a string
  or null, limited to the setting's allowed values where it has a closed set
  (`settings.AllowedValues`, new, returns a copy) and to 128 characters for the
  repo selector, with no other keys. The schema-parity test checks its property
  names against the spec's `SettingsUpdate`. Rejected: a hand-copied list of
  values, which drifts.
- **The endpoint's authorisation.** `PATCH /api/settings` is not session-only,
  and the setting is per account, so an API token may use the tool as it may use
  the endpoint.
- **One mapping.** The endpoint's error classification is extracted
  (`classifySettingsError`) and shared.

## Parity is complete on the Go side

This is the last operation without a tool, so the temporary `pendingMCPTools`
list in the parity test goes: every operation now has a tool or an
`x-agent-excluded` reason, with no third option.

## Capabilities

### Modified Capabilities

- `mcp-endpoint`: the tools cover updating the settings.

## Impact

- `internal/httpserver/mcp_settings_update.go`, `settings.go`, `mcp.go`;
  `internal/settings/settings.go`; `go.mod` (`jsonschema-go` becomes a direct
  dependency).

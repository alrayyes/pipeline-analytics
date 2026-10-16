# Proposal

## Why

A review of the last WebMCP parity pull request found two exclusion reasons that
did not fit their operation, and a parity test that did not check the case its
Go twin does (#696). The exclusion list is only worth having if each reason
says why that operation is out.

## What changes

- `listForgeTokens` and `listCredentials` get reasons about the read itself: the
  list is for the person's own screen, no agent task needs it, and the
  operations beside it are session-only.
- The TypeScript parity test fails when an operation is both a tool and
  excluded, as the Go test already does.
- A dangling sentence in the Go parity test is finished.

## Impact

- Added requirement to `webmcp-support`. The MCP server already fails on an
  operation that is both a tool and excluded (`mcp-endpoint`), so only the
  WebMCP side needed it.

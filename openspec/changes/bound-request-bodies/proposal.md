# Proposal

## Why

No handler limited how much of a request it read, and the spec declared no
length on any request field. One oversized body could use all the memory a
decoder was willing to allocate. Spectral's OWASP ruleset flagged it as API4,
unrestricted resource consumption (#417). Tracked in #429, with the
[OWASP API4 guidance](https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/)
as the source.

## What Changes

- A middleware caps every request body: 64 KiB for the API and the MCP
  endpoint, 2 MiB for the forge webhook receivers. A declared length over the
  cap gets a 413 before anything is read; an undeclared one is cut at the cap.
- The free-text request fields have a `maxLength` in the spec and are checked
  in the handlers (the repo selector in the settings domain) with a 400.
- Every operation that takes a body documents a 413.
- Spectral's string-limit rule is back on for the three request schemas.

## Capabilities

### New Capabilities

- `request-limits`: bounds on what a caller can send.

# Proposal

## Why

Seven e2e specs looped one journey over light and dark: each iteration started
an isolated server, signed in, mocked the network and loaded the page, to run
the same axe scan under a different colour scheme (#614). `testing.md` says one
journey per flow, not one per state. The scan per scheme stays, since contrast
differs; the sign-in, server and page load per scheme go.

## What Changes

- `expectFitsAndPassesAxeInBothSchemes` in `axe.ts` checks the phone-width fit
  once, scans in light, reloads in dark and scans again.
- `overview`, `failures`, `flaky`, `runs`, `branch`, `tabbar` and `error-pages`
  call it from one test each, and the error pages keep one test per width.
- 7 tests fewer on the 14 that looped, with the same 14 scans.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: a requirement that an axe journey covers both colour schemes
  from one sign-in.

## Impact

- `web/tests/e2e/` only.

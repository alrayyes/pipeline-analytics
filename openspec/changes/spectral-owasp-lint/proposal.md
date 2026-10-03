# Proposal

## Why

Redocly checks the OpenAPI description is correct, not that its design is
safe. The account's `api.md` adds Spectral with the OWASP API Security
ruleset for that. It found a real gap (request bodies are unbounded, now
#429) and several rules that don't fit this service (now listed with
reasons). Tracked in #417.

## What Changes

- `.spectral.yaml` extends `@stoplight/spectral-owasp-ruleset`; both packages
  are pinned in the root `package.json`.
- A `spectral (owasp)` CI job and a pre-push hook run it when the `api` group
  changes. Only error-severity findings fail.
- Each rule turned off says why, and points at #429 or #430 where the answer
  is still open. `servers[0]` declares `x-internal: false`.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: the spec is linted for security as well as correctness.

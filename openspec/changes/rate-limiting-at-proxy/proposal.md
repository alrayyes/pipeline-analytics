# Proposal

## Why

Spectral's OWASP ruleset reports a missing rate limit and 429 response. The
service is single-user and normally sits behind a reverse proxy, so the
question was whether it should limit itself. Decided on #430: no, the proxy
does. This records the decision where the next reader looks.

## What Changes

- The README's deployment section says the service doesn't rate-limit, names
  the routes reachable without a session, and points at the proxy.
- The two Spectral exclusions cite the decision instead of the open ticket.
- No code change.

## Capabilities

### New Capabilities

- `deployment`: what the service leaves to its environment.

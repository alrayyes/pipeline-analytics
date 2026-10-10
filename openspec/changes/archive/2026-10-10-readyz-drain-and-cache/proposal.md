# Proposal

## Why

`GET /readyz` checked the database on every request and the server closed its
listener the moment it got SIGTERM, so a router saw no failing readiness before
requests started to drop, and a probe flood reached the database. The account's
`api.md` asks for a cached result and a 503-first shutdown. Tracked in #416.

## What Changes

- `/readyz` reuses its result for a few seconds.
- On SIGTERM `/readyz` answers 503 at once and the server keeps serving for
  `--drain-period` before it closes, then waits at most `--shutdown-timeout`
  for in-flight requests.
- Defaults add up to under the ten second stop grace Docker gives a container.

## Capabilities

### Modified Capabilities

- `http-observability`: readiness caches and drains.

# Proposal

## Why

The image's `HEALTHCHECK` (#316) runs `pipeline-analytics healthcheck`,
which asks `/healthz`. That handler returns 200 unconditionally, so the
container reports `healthy` while its SQLite database is unreadable or the
disk is full: the process answers, nothing else is checked. An operator
looking at `docker ps` can't tell a working instance from one that can't
serve. Tracked in #362.

## What Changes

- New public `GET /readyz`: 200 when the database answers a schema read
  within a second, 503 with a generic error body otherwise. The cause is logged, not
  returned.
- `pipeline-analytics healthcheck` asks `/readyz` instead of `/healthz`, so
  the Dockerfile's `HEALTHCHECK` reports readiness.
- `/healthz` is unchanged: still a plain liveness probe that stays 200 with
  the database down, so an orchestrator never restarts a process over a
  dependency it can't fix.
- `/healthz` and `/readyz` are both described in `openapi.yaml` (`/healthz`
  wasn't).
- Neither probe is request-logged, as `/healthz` already isn't.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `http-observability`: the request-log exemption widens from the liveness
  endpoint to the liveness and readiness endpoints, and a readiness
  endpoint requirement is added.

## Impact

- Backend: `Deps.Ready`, a handler and two middleware exemptions in
  `internal/httpserver`; `main.go` wires a schema read and points the
  `healthcheck` subcommand at `/readyz`.
- Contract: two paths added to `openapi/openapi.yaml`. The SDK's contract
  test discovers operations by reflection, so the regeneration PR needs no
  hand edit.
- `Dockerfile` comment and the README's Docker section.
- No config, no migration, no new dependency.

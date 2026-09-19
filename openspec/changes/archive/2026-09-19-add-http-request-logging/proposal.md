# Proposal

## Why

`internal/httpserver.New()` wraps its mux with only `requireSession` —
there's no request-logging middleware at all. The default
`--log-level` is `info`, and every existing log call in this codebase
(webhook receipt, ingestion reconciliation) is at `Debug`, so a
default deployment currently logs almost nothing about what's hitting
the server: not a successful request, not a failed one, not a rejected
auth attempt.

## What Changes

- A request-logging middleware wraps the whole handler tree (outside
  `requireSession`, so a rejected/unauthenticated request is logged
  too), emitting one structured `slog` line per request after it
  completes: method, path, status, duration, remote address.
- Log level varies by outcome, not a flat `Info` for everything: 5xx →
  `Error`, 4xx → `Warn`, everything else → `Info` — matching this
  codebase's existing convention of reserving `Error` for real
  failures rather than drowning it in routine traffic.
- `/healthz` is excluded from request logging (or logged at `Debug`)
  — it's polled continuously by whatever's checking liveness and adds
  no actionable signal at `Info`.
- The middleware never logs the `Cookie` or `Authorization` header,
  the raw query string, or request/response bodies — only the fields
  listed above.

## Capabilities

### New Capabilities

- `http-observability`: request-level logging behavior for the HTTP
  server, cross-cutting every endpoint rather than belonging to any
  single existing capability.

### Modified Capabilities

(none)

## Impact

- Backend: new middleware in `internal/httpserver` (likely
  `internal/httpserver/logging.go`), wired into `New()` around the
  existing `requireSession(mux)` call.
- No API contract change — this is server-side observability, not a
  response shape or new endpoint.
- Not included in this change: panic recovery (same middleware layer,
  a related but separate gap — no `recover()` exists anywhere in
  `internal/httpserver` today). Flagged as a possible follow-up, not
  folded in here to keep this change to one concern.

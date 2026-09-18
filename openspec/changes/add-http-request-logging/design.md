# Design

## Context

See `proposal.md` - Why. Relevant existing state:

- `internal/httpserver/server.go`'s `New()` returns
  `requireSession(deps.AuthStore, mux)` - a single middleware layer,
  auth only. Adding a second layer means wrapping that same return
  value, not restructuring the mux itself.
- Every existing `slog` call in this codebase (`internal/ingestion`,
  `internal/httpserver/webhooks.go`) uses `*Context` variants
  (`slog.DebugContext`, `slog.ErrorContext`) with plain, flat
  attribute keys (`"forge"`, `"repo"`, `"error"`) - no OpenTelemetry-
  style dotted names. This middleware follows the same convention.
- `cmd/pipeline-analytics/main.go` sets the default logger via
  `slog.NewTextHandler` at whatever level `--log-level` resolves to
  (default `info`). The middleware doesn't change that setup - it logs
  through the existing default logger like everything else does.
- `net/http`'s `http.ResponseWriter` doesn't expose the status code a
  handler wrote; capturing it needs a small wrapping type.

## Goals / Non-Goals

**Goals:**

- Every request produces exactly one log line, including ones
  `requireSession` rejects before a handler ever runs.
- No secret-bearing header, query string, or body content ever
  reaches a log line.

**Non-Goals:**

- Panic recovery. Related (same middleware layer), but a separate
  concern from request logging - see proposal.md's Impact section.
- A request ID / trace correlation mechanism. Worth having
  eventually, but this app has no distributed tracing to correlate
  against yet, and one process's stderr log doesn't need a
  cross-service correlation ID to stay readable - revisit if that
  changes.
- Sampling or rate-limiting the log output. This app's traffic (a
  solo developer's dashboard) doesn't approach a volume where request
  logging itself becomes a cost.

## Decisions

**Middleware wraps `requireSession(mux)` from the outside, not the
mux directly.** `New()` becomes
`requestLogger(requireSession(deps.AuthStore, mux))`. This is why a
401 from `requireSession` still gets logged - the logging middleware
sees the response `requireSession` produced, not just what the mux's
handlers produce.

**Status capture via a minimal `http.ResponseWriter` wrapper**, not a
third-party middleware library. A small unexported
`statusRecorder` struct embedding `http.ResponseWriter` and overriding
`WriteHeader` to record the status (defaulting to 200 if
`WriteHeader` is never called explicitly, matching `net/http`'s own
default) is the standard, dependency-free way to do this - not worth
adding a router/middleware framework for one wrapper type.

**`/healthz` is excluded outright, not logged at `Debug`.** It's
polled continuously by whatever checks liveness (a container
orchestrator, an uptime check) and produces zero actionable signal
even at `Debug` - excluding it entirely is simpler than a level
override that still writes (and someone eventually enables) noise.
Every other path, including the webhook receivers and the SPA static
handler, is logged the same way.

**Level thresholds match HTTP status *classes*, not individual
codes.** 5xx -> `Error`, 4xx -> `Warn`, else `Info` - the standard,
widely-recognized breakdown (matches how most web frameworks' access-
log integrations classify outcomes), and simple enough to implement
as one `switch` on `status / 100`.

## Risks / Trade-offs

- **A 4xx logged at `Warn` includes routine, expected client
  behavior** (a bad request body, a 404 for an untracked repo) ->
  could read as noisier than intended at `Warn`. Accepted: `Warn` for
  "the client did something the server rejected" is the standard
  reading of a 4xx, and it's still one level below `Error`, which
  stays reserved for the server's own failures.
- **The static/SPA handler (`mux.Handle("/", staticHandler(...))`)
  logs a line per asset request** (JS chunks, images), which could
  dominate the log volume compared to API calls. Accepted for now
  given this app's scale; revisit (e.g. excluding non-`/api/`,
  non-`/webhooks/` paths) if it turns out to matter in practice rather
  than pre-optimizing for a volume this app doesn't have.

## Migration Plan

1. Add the `statusRecorder` wrapper and the logging middleware in
   `internal/httpserver`.
2. Wire it into `New()` around `requireSession(mux)`.
3. Verify manually (or via an integration test) that a rejected
   (401), a client-error (4xx), a server-error (5xx), and a normal
   (2xx) request each produce a log line at the right level, and that
   `/healthz` produces none.

No rollback concern: this only adds log output, nothing behavioral
changes for a client.

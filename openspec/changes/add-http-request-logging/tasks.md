# Tasks

## 1. Middleware

- [x] 1.1 Add a `statusRecorder` `http.ResponseWriter` wrapper in
      `internal/httpserver` that captures the written status
      (defaulting to 200), with a unit test covering both an explicit
      `WriteHeader` call and none
- [x] 1.2 Add the request-logging middleware: one `slog` line per
      request (method, path, status, duration, remote address), level
      chosen by status class (5xx Error, 4xx Warn, else Info), with
      unit tests for all three level cases
- [x] 1.3 Exclude `/healthz` from logging entirely, with a test
      verifying no log record is produced for it
- [x] 1.4 Verify no test asserts on a log line containing a `Cookie`,
      `Authorization`, or query-string value (regression guard for the
      "no sensitive data" requirement)

## 2. Wiring

- [x] 2.1 Wrap `New()`'s existing `requireSession(deps.AuthStore,
      mux)` with the new middleware, and verify a 401 from
      `requireSession` itself produces a log line

## 3. Ship it

- [x] 3.1 Open a pull request with `Closes #196` and verify CI passes

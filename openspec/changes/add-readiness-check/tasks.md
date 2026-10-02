# Tasks

## 1. Contract

- [x] 1.1 Describe `/healthz` and `/readyz` in `openapi/openapi.yaml`,
      linted with Redocly

## 2. Readiness endpoint

- [x] 2.1 Failing tests: 200, 503 without leaking the cause, no session
      needed, probe deadline, not request-logged, liveness unaffected
- [x] 2.2 `Deps.Ready`, the handler, and the two middleware exemptions
- [x] 2.3 Wire a schema read (`dbReady`) in `main.go`; tests that a closed
      connection and a non-database file fail it, and that the real
      handler's `/readyz` goes 503 with the database

## 3. Container check

- [x] 3.1 Point the `healthcheck` subcommand at `/readyz`; test that it asks
      for that path
- [x] 3.2 Update the Dockerfile comment and the README's Docker section
- [x] 3.3 Verify against a built image: `healthy` with a good database,
      `unhealthy` with the file unreadable

## 4. Wrap up

- [ ] 4.1 Archive this change once merged

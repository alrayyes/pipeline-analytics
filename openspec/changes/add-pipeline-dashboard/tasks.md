## 1. Foundation (complete)

- [x] 1.1 Go server scaffold: cobra/viper `serve` command, SQLite connection (modernc.org/sqlite), graceful shutdown, `/healthz` -- `go build ./... && go test ./...` passes (github.com/alrayyes/pipeline-analytics#8)
- [x] 1.2 CI pipeline: build/vet/test/race/cover, `go mod tidy -diff`, govulncheck, golangci-lint v2, all required status checks -- green on `main`
- [x] 1.3 OpenAPI spec for every v1 endpoint, Redocly-linted and lint-gated in CI -- `bunx @redocly/cli lint` passes (github.com/alrayyes/pipeline-analytics#12)

## 2. forge-ingestion

- [x] 2.1 SQLite schema/migrations for repos, runs, jobs, steps, and encrypted credentials -- verify a fresh `:memory:` database migrates cleanly
- [x] 2.2 Encrypted-at-rest credential storage for repo-scoped PATs -- verify a stored token is never returned in plaintext from any read path
- [x] 2.3 `POST /api/repos` handler: register a repo, create its webhook via `google/go-github` (GitHub) or the Gitea SDK (Forgejo) -- verify against forge-ingestion/spec.md's registration scenarios using `httptest.Server` fakes for each forge's REST API
- [x] 2.4 Webhook receiver handlers (`POST /webhooks/github`, `POST /webhooks/forgejo`): HMAC signature verification, run/job/step upsert -- verify the valid-signature and invalid-signature scenarios
- [x] 2.5 Hourly reconciliation poller: GitHub conditional-GET (ETag/If-None-Match, no rate-limit cost on 304), Forgejo REST poll, backfill on newly tracked repo -- verify the backfill and missed-webhook scenarios
- [x] 2.6 Wire ingestion into `cmd/pipeline-analytics` (reconciliation scheduler started, webhook routes registered) -- verify `pipeline-analytics serve` exposes the new routes and the scheduler runs

## 3. dashboard-auth

- [x] 3.1 WebAuthn registration ceremony (`go-webauthn/webauthn`): options + verify endpoints, single-user enforcement -- verify the first-run and second-account-rejected scenarios
- [x] 3.2 WebAuthn login ceremony: options + verify endpoints, session issuance -- verify the successful-login and failed-assertion scenarios
- [x] 3.3 Session middleware gating every route except login/registration and the webhook receivers -- verify an unauthenticated request is denied (JSON 401, not a server-side redirect -- every route is a client-rendered SPA page per design.md's SPA-fallback decision, so the client router redirects to login after seeing 401, not the server)

## 4. pipeline-metrics

- [ ] 4.1 Duration percentile computation (p50/p90) per pipeline and per step over a rolling window -- verify against a fixture run history
- [ ] 4.2 Queue-time vs. execution-time split per job -- verify the two are reported as separate values
- [ ] 4.3 Failure-rate trend per pipeline and per step -- verify step-level rate is independent of the pipeline's overall rate
- [ ] 4.4 Flaky-step detection (inconsistent pass/fail vs. consistently broken) -- verify a flaky step is distinguishable from a consistently-failing one
- [ ] 4.5 Slowest-step ranking -- verify steps are ordered by duration contribution, highest first
- [ ] 4.6 Health-status computation combining failure rate, duration regression, and flakiness -- verify the healthy/unhealthy scenarios from pipeline-metrics/spec.md
- [ ] 4.7 Actions-minutes usage tracking per workflow -- verify usage is reported broken down by workflow
- [ ] 4.8 Wire `GET /api/pipelines`, `GET /api/pipelines/{id}`, `GET /api/pipelines/{id}/steps`, `GET /api/repos/{id}/usage` handlers against `openapi/openapi.yaml`

## 5. dashboard-ui

- [x] 5.1 SvelteKit project scaffold under `web/` (bun, TypeScript strict), built output embedded into the Go binary via `go:embed` -- verified via `bun run build` output served by the running binary, plus a footer + release-history page (github.com/alrayyes/pipeline-analytics#16). shadcn-svelte/LayerChart intentionally not installed yet -- nothing to chart until pipeline-metrics (#4) exists; installed alongside 5.3, their first real consumer.
- [ ] 5.2 WebAuthn login/registration UI flow -- verify end-to-end login against a running server
- [ ] 5.3 Repo/pipeline overview page (installs shadcn-svelte + LayerChart) -- verify every tracked pipeline's health status is visible without opening it individually
- [ ] 5.4 Pipeline detail view: duration trend and failure-rate trend charts -- verify a regressed trend is visible in the chart, not only the current aggregate
- [ ] 5.5 Step breakdown view: duration ranking, queue/exec split, failure rate, flaky callout -- verify flaky and consistently-failing steps are visually distinguishable
- [ ] 5.6 Deep links from every displayed run/job/step to its page on the originating forge -- verify the link opens the correct forge URL
- [ ] 5.7 Usage view: runner-minutes per workflow -- verify displayed broken down by workflow

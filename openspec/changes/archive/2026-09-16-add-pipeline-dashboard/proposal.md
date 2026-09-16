## Why

A solo developer running CI across GitHub Actions and Forgejo Actions has no unified, historical view of pipeline health: each forge's own UI shows recent runs in isolation, with no cross-forge view, no persisted trend history, and no explicit flaky-step signal. This makes regressions (a step getting slower, a step starting to fail intermittently) invisible until they're already a recurring annoyance.

## What Changes

- New self-hosted Go service (SQLite storage) that ingests GitHub Actions and Forgejo Actions run/job/step data via webhooks (primary) with hourly REST reconciliation (conditional-GET on GitHub to avoid rate-limit cost; no GraphQL).
- New SvelteKit dashboard (shadcn-svelte + LayerChart), embedded into the Go binary at build time, single-binary deploy via Docker on a VPS.
- WebAuthn-based login for the dashboard itself, distinct from the pasted, repo-scoped GitHub/Forgejo PATs used to read Actions data and manage webhooks.
- Computed pipeline/step metrics: health score, duration trend (p50/p90), queue-time vs. execution-time split, failure-rate trend, flaky-step detection (inconsistent pass/fail vs. consistently broken), slowest-step ranking, Actions-minutes usage tracking.
- Diagnosis-first UX: every flagged issue deep-links to the exact run/step on GitHub or Forgejo. No write access to either forge (no re-run-from-dashboard, no forge-side mutation) and no outbound alerting in v1.
- License: AGPL-3.0.

## Capabilities

### New Capabilities
- `forge-ingestion`: Registers tracked repos against pasted GitHub/Forgejo PATs, creates and verifies webhooks on those repos, receives and stores workflow run/job/step events, and runs hourly reconciliation polling (conditional-GET on GitHub) to backfill history and catch missed deliveries.
- `pipeline-metrics`: Derives health score, duration percentiles, queue-vs-exec split, failure-rate trend, flaky-step detection, slowest-step ranking, and Actions-minutes usage from ingested run data.
- `dashboard-auth`: WebAuthn registration and login for the dashboard's single solo-developer user.
- `dashboard-ui`: SvelteKit dashboard presenting per-repo/per-pipeline health, drill-down into step-level timing and failures, and deep links to the originating GitHub/Forgejo run.

### Modified Capabilities
None — this is a new project with no existing specs.

## Impact

- New repository (`pipeline-analytics`), currently an empty OpenSpec scaffold with no code and not yet a git repo.
- New Go API service, SQLite database, webhook receiver endpoint (needs a public HTTPS-reachable URL for GitHub webhook delivery; Docker on a VPS handles this).
- New SvelteKit frontend, built and embedded into the Go binary.
- Stored credentials: encrypted-at-rest GitHub and Forgejo PATs, WebAuthn credential records.
- External dependencies: `google/go-github` (GitHub REST client), Gitea SDK (`code.gitea.io/sdk/gitea`, used against Forgejo — Actions-endpoint coverage to be verified at implementation time), a Go WebAuthn library (`go-webauthn/webauthn`), shadcn-svelte, LayerChart.

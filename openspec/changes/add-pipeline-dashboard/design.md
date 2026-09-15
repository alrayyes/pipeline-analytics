## Context

New project, no existing code. See proposal.md - Why for motivation. Single user (the solo developer), self-hosted, deployed via Docker on a VPS with a public HTTPS URL (required for GitHub webhook delivery; Forgejo may be same-instance or elsewhere).

## Goals / Non-Goals

**Goals:**
- Single-binary-friendly deployment: one Go process serving both the API and the built SvelteKit frontend.
- Minimize GitHub REST rate-limit consumption without adding a second API paradigm (no GraphQL).
- Keep the forge integration read-mostly: only a webhook-management write (create/verify webhook) against either forge, nothing else.
- Every HTTP endpoint is designed in a committed OpenAPI description before its handler is written (per this account's `api.md` convention) -- the spec for a capability's endpoints lands as its own reviewed step before that capability's handler code, not alongside it.

**Non-Goals:**
- Multi-user/multi-tenant support (single WebAuthn user in v1).
- Re-running or otherwise mutating pipelines from the dashboard.
- Outbound alerting/notifications (email, Slack, webhook-out) in v1.
- GitHub App / Forgejo OAuth app registration flow (deferred; pasted PAT is v1's auth model).

## Decisions

### Backend: Go + SQLite
Go compiles to a single static binary that can `go:embed` the built SvelteKit frontend, so deployment is "build once, run one binary" - a strong fit for a solo-dev self-hosted tool. SQLite is sufficient for one user's historical run data and needs no separate database process. Alternative considered: TypeScript (Node/Bun) API - rejected for this project because it requires a runtime and `node_modules` on the host in addition to the frontend build, adding an operational dependency Go avoids; TypeScript remains the frontend language regardless, via SvelteKit.

### Frontend: SvelteKit, not Astro
The dashboard is almost entirely dynamic, login-gated, per-viewer state - the opposite of Astro's mostly-static-with-islands model. Using Astro here would mean wrapping a single giant interactive island in a framework built for content sites. SvelteKit is built for this shape of app (client routing, layouts, form actions for the WebAuthn ceremony) with no static-site baggage.

### UI components: shadcn-svelte + LayerChart
shadcn-svelte is copy-paste-owned (not an opaque dependency) and built on Bits UI headless primitives, so accessibility (ARIA, keyboard nav, focus management) is handled rather than hand-rolled. LayerChart (composable Svelte chart components on D3/Layer Cake) is the charting library shadcn-svelte's own chart component is built on, so the two integrate without extra glue. Both are MIT-licensed, compatible with this project's AGPL-3.0 license. Alternative considered: Chart.js/ApexCharts - rejected as harder to reskin into a consistent, accessible, light/dark-aware design system than a composable low-level library.

### Ingestion: webhooks primary, REST reconciliation fallback, no GraphQL
Webhooks (`workflow_run`/`workflow_job` events) carry the real-time load at zero ongoing API cost. Reconciliation polling runs hourly per tracked repo to backfill new repos and catch missed deliveries. On GitHub, reconciliation uses conditional requests (`If-None-Match` against the response `ETag`): an authenticated request that returns `304 Not Modified` does not count against the 5,000 req/hour primary rate limit, provided the same token is used consistently. This was chosen over GitHub's GraphQL API - GraphQL batching would reduce request count, but conditional REST achieves the actual goal (minimizing rate-limit consumption) without introducing a second API client/paradigm alongside the REST client webhook management already needs. Forgejo has no GraphQL API (REST/Swagger only) and no default rate limit, so this question doesn't apply there.

### Forge SDKs
`google/go-github` for GitHub (handles conditional-request plumbing). The Gitea SDK (`code.gitea.io/sdk/gitea`) for Forgejo, since Forgejo's REST API is Gitea-derived. Actions-specific endpoint (workflow runs/jobs/steps) coverage in the Gitea SDK should be verified early in implementation - Actions is newer surface, and community forks of the SDK exist specifically to patch Forgejo-specific gaps, which may mean falling back to raw REST calls for those endpoints.

### Forge authentication: pasted repo-scoped PAT
The user pastes a fine-grained GitHub PAT (scoped to the specific repos being tracked, with Actions-read and webhook-management permissions) and a Forgejo access token, stored encrypted at rest. Chosen over a GitHub App / Forgejo OAuth app registration flow: the app-based approach is more secure and avoids manual rotation, but its added build cost (app manifest, install callback, token refresh) isn't justified for a tool with exactly one user. This is a deliberate v1 scope boundary, not a rejected idea - see Open Questions.

### Auth: WebAuthn only, no password path
Matches the proposal's stated requirement directly. No password fallback is offered, since a solo-dev tool with one user has no forgotten-password support burden to justify one, and passkey-only reduces the credential-phishing surface for a dashboard reachable at a public URL.

### Actionability: diagnose + deep-link only
The dashboard identifies the specific failing/slow step and links to it on the originating forge, rather than acting on the user's behalf (re-run, alerting). This keeps forge write scope limited to webhook management, avoids building a notification-delivery subsystem in v1, and keeps the token-scope/attack-surface footprint minimal for a public-URL deployment.

### API design: OpenAPI first, Redocly-linted
Every endpoint -- repo registration, WebAuthn ceremonies, the webhook receivers, and the dashboard's own data APIs -- is hand-written into `openapi/openapi.yaml` and reviewed before its handler exists, per this account's standing `api.md` convention. Linted with Redocly CLI (`recommended-strict`, since the spec is new), one `redocly.yaml` at the repo root. Server stubs/typed clients may be generated from it later; the spec itself is never generated from annotations. A capability's spec PR lands and is reviewed on its own, then its handler PR follows -- not combined into one PR that changes both.

### Server entrypoint: cobra + viper
The binary's command surface is `spf13/cobra` (a `serve` command now; room for a one-off admin/migration command later without restructuring), with `spf13/viper` for configuration -- flags, environment variables, and an optional config file, in that precedence, bound via `viper.BindPFlag`. A `Validate() error` on the config struct runs once after `viper.Unmarshal`, so a bad value fails at startup rather than wherever it's first read.

### Architecture: hexagonal, ports owned by the domain
Domain logic (health-status computation, flaky-step detection, duration/failure-rate aggregation) sits at the centre, importing neither `net/http` nor `database/sql`. It's tested with fakes. Ports -- a run/job/step store, a forge client -- are interfaces defined in the domain package that needs them; adapters (the SQLite-backed store, the GitHub/Forgejo REST clients built on `google/go-github`/the Gitea SDK, the HTTP handler tree) implement those ports at the edges and are wired together at the composition root in `cmd/pipeline-analytics/main.go`. Package layout follows `internal/<capability>` (`internal/ingestion`, `internal/metrics`, `internal/auth`, `internal/httpserver`), not a `domain/application/infrastructure` tree -- there's no DI framework, dependencies are passed explicitly from `main`.

### Testing: outside-in, testify, containers only where a real service exists
Per this account's `testing.md`/`go-test.md`: the outer (acceptance/integration) test goes in first against the interface a caller wants, unit tests fill in branches and edge cases once that's red for the right reason. Unit tests use testify (`require`/`assert`), external `_test` packages, table-driven cases with `t.Parallel()`. SQLite is embedded, not a separate service, so `testcontainers-go` doesn't apply to it; it's reserved for a genuine external dependency if one is ever added. GitHub/Forgejo API interaction is tested against an `httptest.Server` fake implementing the subset of each forge's REST API this project actually calls (webhook creation, run/job listing, conditional-GET semantics) -- there's no real forge to containerize, and a fake gives control over rate-limit/304 responses a live API wouldn't.

### Frontend tooling: bun
The SvelteKit frontend uses bun as the runtime, package manager, bundler, and test runner (`bun run`, `bun test`, `bun build`), not npm/node -- this account's default per `javascript.md`. Lockfile is `bun.lock`, committed. Linted with Biome (JS/TS/JSON/CSS in one pass); Prettier only for the Markdown/YAML gaps Biome doesn't cover.

### Containerization: hardened multi-stage Dockerfile
Multi-stage build (Go toolchain image to compile, a minimal runtime base -- alpine or distroless -- for the artifact), non-root `USER`, base image pinned by digest, bound to a high port (8080) rather than granting a low-port capability. Hadolint runs against the Dockerfile in the same hook/CI step as everything else.

### Dependency hygiene and releases
Dependabot watches both ecosystems (Go, and bun/npm once the frontend's `package.json` exists) once each manifest lands. `govulncheck ./...` (Go) and `bun audit` (frontend) run in CI on every push, separately from the update bot. Since this repo is GitHub-hosted, releases use **release-please** (reads Conventional Commits, keeps a release PR open) plus **goreleaser** (cross-compiles and attaches the binary once that PR merges and tags) -- not semantic-release, which is this account's Forgejo-side equivalent. `RELEASE_TOKEN` is already provisioned as a repo secret.

## Risks / Trade-offs

- [Pasted PAT model requires manual rotation and has no automatic revocation path] -> Mitigated by encrypting tokens at rest and scoping them to only the tracked repos; acceptable for a single-user tool, revisit if this ever becomes multi-user.
- [Gitea SDK's Forgejo Actions endpoint coverage is unverified] -> Verify during early implementation; fall back to raw REST calls against Forgejo's documented Actions API where the SDK is missing coverage.
- [Public-URL deployment with no forge write access still exposes a webhook receiver endpoint to the internet] -> Mitigated by per-repo webhook signature verification (HMAC) rejecting any unverified delivery before it touches stored data.
- [Health-score formula (how duration regression, failure rate, and flakiness combine into one status) is not fully specified] -> See Open Questions; the individual signals are each specified in pipeline-metrics/spec.md and can ship independently of a single combined score if needed.
- [OpenAPI-spec-first adds a design/review step before any handler for a capability can be written] -> Accepted cost for a stable, reviewed contract; the spec PR is usually small relative to the handler PR that follows it.
- [cobra/viper is real ceremony for a binary that, in v1, has exactly one command] -> Accepted for consistency with this account's standing Go convention and to leave room for a later admin/migration command with no restructuring.

## Open Questions

- Exact health-score formula/weighting (how duration regression, failure-rate trend, and flakiness combine into a single healthy/unhealthy status) - can be tuned after implementation without changing the underlying per-signal specs.
- Data retention window for historical run data (keep indefinitely vs. prune after N days) - doesn't change the ingestion or metrics behavior, only storage growth; can be decided and added as a pruning job later.
- Whether GitHub App / Forgejo OAuth app support is ever added post-v1 - explicitly deferred, not decided against.

# Architecture

pipeline-analytics is a single Go binary that serves a REST API and
`go:embed`s a built SvelteKit frontend, backed by an embedded SQLite
database. It ingests GitHub/Forgejo Actions run history via webhooks
(primary) with a reconciliation poll (fallback), computes each tracked
pipeline's health status and slowest-step ranking live on every request,
and authenticates its single account with WebAuthn only.

Domain logic (`internal/metrics`, `internal/ingestion`, `internal/auth`)
follows a hexagonal shape: ports are interfaces the domain package
itself defines, adapters (the SQLite-backed stores, the GitHub/Forgejo
REST clients, the HTTP handler tree) implement them at the edges, and
everything is wired together explicitly in `cmd/pipeline-analytics/main.go`
-- no DI framework, no `domain/application/infrastructure` tree.

## Why it's built this way

Each significant decision -- one that would cost real rework to reverse,
or where an obvious alternative was already tried and rejected -- has its
own record in [`docs/adr/`](docs/adr/):

- [0001 - Go backend with an embedded SQLite database](docs/adr/0001-go-and-sqlite-backend.md)
- [0002 - WebAuthn-only authentication, no password path](docs/adr/0002-webauthn-only-authentication.md)
- [0003 - Webhook-primary ingestion with a polling fallback](docs/adr/0003-webhook-primary-poll-fallback-ingestion.md)
- [0004 - Pipeline health is recomputed live, nothing persisted](docs/adr/0004-live-recomputed-health-status.md)

A decision that changes gets a _new_, numbered record that supersedes the
old one, rather than an edit -- the old record's reasoning stays as it
was, since it's exactly what the next person needs to see wasn't
arbitrary.

This isn't the same artifact as an OpenSpec change's own `design.md`
(`openspec/changes/`) -- those are one change's working papers, archived
once it ships. An ADR here is the durable, standalone answer to "why,"
meant to still be the first place someone looks well after the change
itself is forgotten.

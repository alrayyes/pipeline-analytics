# 1. Go backend with an embedded SQLite database

## Status

Accepted

## Context

This is a self-hosted tool for a single user tracking their own repos --
not a multi-tenant service. The backend needs to serve both a REST API
and the built SvelteKit frontend, store historical run/job/step data for
one account, and deploy as simply as possible onto a VPS.

Alternative considered: a TypeScript (Node/Bun) API, since the frontend
is already TypeScript. Rejected because it requires a runtime and
`node_modules` on the host in addition to the frontend build, adding an
operational dependency a compiled Go binary avoids.

## Decision

The server is a single Go binary that `go:embed`s the built SvelteKit
frontend, so deployment is "build once, run one binary" behind Docker or
systemd. Storage is SQLite, embedded in-process rather than a separate
database server -- sufficient for one account's historical run data,
with nothing else to run, back up, or network to.

## Consequences

- Deployment is one binary plus one file (the SQLite database), matching
  the "one Docker container, one volume" shape the README documents.
- There's no path to a separate read replica or horizontal scale-out of
  storage without a real migration -- acceptable because this is
  explicitly single-account, single-writer software, not a target this
  project is trying to hit.
- Go's `internal/` package boundary plus a hexagonal ports/adapters split
  (see `internal/metrics.Store`, `internal/ingestion`) keeps the SQLite
  dependency behind an interface each domain package owns, so swapping
  the storage engine later -- if this project's shape ever changes --
  is a new adapter, not a rewrite of domain logic.

## Citations

- `openspec/changes/archive/2026-09-16-add-pipeline-dashboard/design.md`,
  "Backend: Go + SQLite" and "Architecture: hexagonal, ports owned by the
  domain."

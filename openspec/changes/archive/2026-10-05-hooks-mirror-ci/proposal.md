# Proposal

## Why

`lefthook.yml` leaves out checks CI runs: hadolint, `goreleaser check`,
Redocly and the `go.mod` formatting diff. A push can pass every hook and
still go red. On a machine with no Docker cache directories, Docker also
creates them as root, so the first Go hook fails with permission denied.
Tracked in #420.

## What Changes

- `pre-push` gains hadolint, `goreleaser check`, Redocly and the `go.mod`
  formatting check, each guarded by the CI group that runs it.
- Every Go hook creates its cache directories before Docker mounts them.
- `lefthook.yml` sets `output: [failure]`.
- `lefthook.yml` names the checks that stay CI-only (`govulncheck`,
  `bun audit`) and why.
- A test in the `changes` job fails when a CI check loses its hook.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: local hooks cover the same checks as CI.

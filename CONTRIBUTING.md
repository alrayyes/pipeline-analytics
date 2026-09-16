# Contributing

This is a personal, solo-maintained project, but the setup here is real and
usable if you want to run it locally, poke at the code, or send a fix.

## Toolchain

- Go, version from [`go.mod`](go.mod) (currently 1.27+).
- [bun](https://bun.sh) 1.3.x — not 1.4+. Bun 1.4 defaults to a
  `bun.lock` format Dependabot's own bun updater can't read yet; see the
  comment in [`web/package.json`](web/package.json).
- [golangci-lint](https://golangci-lint.run) v2, pinned version in
  [`.github/workflows/ci.yml`](.github/workflows/ci.yml).
- [Playwright](https://playwright.dev) needs a Chromium download the first
  time you run the E2E suite: `cd web && bunx playwright install --with-deps chromium`.
- Docker — used to build or test the container image, and to run every Go
  git hook below against a pinned `go`/`golangci-lint`, not whatever's on
  your `PATH`.

## Building

```sh
(cd web && bun install --frozen-lockfile && bun run build)
go build -o pipeline-analytics ./cmd/pipeline-analytics
```

The frontend has to build first — `web/vite.config.ts`'s adapter writes
straight into `internal/webassets/dist`, which the Go binary embeds via
`go:embed`.

## Testing and linting

Go:

```sh
go build ./...
go vet ./...
go test -race -cover ./...
go mod tidy -diff
golangci-lint run ./...
```

Frontend (from `web/`):

```sh
bun run check    # svelte-check
bun run lint     # biome
bun run build
bun audit
bun run test:e2e # Playwright, builds the real binary and runs against it
```

OpenAPI spec:

```sh
bunx @redocly/cli lint
```

Run the same commands CI runs — [`.github/workflows/ci.yml`](.github/workflows/ci.yml)
is the source of truth if anything here drifts from it.

## Local hooks

`bun install` at the repo root (`prepare` script) installs
[lefthook](https://github.com/evilmartians/lefthook)'s git hooks; run
`bunx lefthook install` yourself if you skipped that step. What each hook
runs, straight from [`lefthook.yml`](lefthook.yml):

- **`pre-commit`** (staged files only, fixes and restages): `gofmt`,
  `go mod edit -fmt go.mod`, `biome check --write` for `web/`.
- **`commit-msg`**: [commitlint](https://commitlint.js.org) against
  `@commitlint/config-conventional`.
- **`pre-push`** (whole tree, never writes): `go vet`, `go test -race
  -cover`, `go mod tidy -diff`, `golangci-lint run` -- the Go commands run
  through Docker (`golang:<go.mod's version>-bookworm`,
  `golangci/golangci-lint:<CI's pin>`) so the version checking your push
  is always the one the repo declares, not whatever your package manager
  last updated -- plus `bun run check`, `bun run lint`, and a scoped
  `docker build` when the Dockerfile or Go sources changed.

No hook reaches for a linter CI doesn't also run.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/):
`type(scope): short description`, subject under 50 characters, lowercase,
no trailing period. Types: `feat`, `fix`, `docs`, `style`, `refactor`,
`perf`, `test`, `build`, `ci`, `chore`, `revert`. A breaking change gets a
`BREAKING CHANGE:` footer — [release-please](https://github.com/googleapis/release-please)
reads these to compute the next version.

## Workflow

- Everything lands through a pull request — branch, push, open a PR against
  `main`.
- Spec-first for any new HTTP endpoint: [`openapi/openapi.yaml`](openapi/openapi.yaml)
  gets the addition, reviewed on its own, before the handler.
- CI has to be green (build/vet/test, golangci-lint, redocly-lint, frontend,
  e2e, docker build — all required checks on `main`) before a PR merges.
- The project's current capabilities are specced under
  [`openspec/specs/`](openspec/specs/); a new one goes through
  [OpenSpec](https://github.com/Fission-AI/OpenSpec)'s
  `openspec/changes/` proposal flow before it's implemented.

## Releases

[release-please](https://github.com/googleapis/release-please) computes the
next version from Conventional Commits on `main` and keeps a release pull
request open; merging it tags the release, and
[goreleaser](https://goreleaser.com) cross-compiles the binaries and builds
the multi-arch Docker image. Nobody picks a version by hand.

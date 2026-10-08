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

`bun run build` and `bun run dev` first run `web/scripts/generate-releases.ts`,
which turns `CHANGELOG.md` into `releases.json` for the release history page.
That file is generated and not committed; edit the changelog, not the JSON.

The dashboard also registers three read-only
[WebMCP](https://github.com/webmachinelearning/webmcp) tools
(`list_pipelines`, `get_pipeline`, `get_repo_usage`) on mount, so an
in-browser agent sharing the page can read the same data it renders. This
is feature-detected (`web/src/lib/webmcpTools.ts`) and safe to ignore —
WebMCP is a pre-stable draft with no browser shipping it stably yet, and
registration is a no-op everywhere else.

## Testing and linting

Go:

```sh
go build ./...
go vet ./...
go test -race -cover ./...
go mod tidy -diff
golangci-lint run ./...
```

The code that decodes data we don't control has fuzz tests: the webhook
signature and event decoders, the GitHub and Forgejo response decoders, the
webhook endpoint and the MCP endpoint. `go test` replays their seeds and any
committed crasher. To hunt for new inputs, run one for a while:

```sh
go test ./internal/ingestion -run '^$' -fuzz '^FuzzProcessGitHubEvent$' -fuzztime 1m
```

The `Fuzz` workflow does that for every target weekly, and on demand from the
Actions tab. A crash uploads the input Go wrote under `testdata/fuzz/`, and committing
that file with the fix keeps it as a regression test.

Frontend (from `web/`):

```sh
bun run check          # svelte-check
bun run lint           # biome
bun run test           # unit tests (bun:test)
bun run test:coverage  # same, plus a coverage report and `junit.xml`
bun run test:mutation  # mutation testing (Stryker), fails on a surviving mutant
bun run build
bun audit --ignore GHSA-hp3w-g68c-fv3c
bun run test:e2e        # Playwright, builds the real binary and runs against it
./scripts/lighthouse.sh # Lighthouse CI on /login and the signed-in pages, builds the real binary too
```

`test:mutation` runs against
[`@hughescr/stryker-bun-runner`](https://github.com/hughescr/stryker-bun-runner),
the actively maintained one of the two community `Stryker` runners for Bun
(no official one exists yet) -- see `stryker.conf.mjs` for why.

OpenAPI spec:

```sh
bunx @redocly/cli lint
bunx spectral lint openapi/openapi.yaml
```

Redocly checks the spec is correct; Spectral adds the OWASP API security
rules. Only its error-severity findings fail, and
[`.spectral.yaml`](.spectral.yaml) gives the reason for every rule it disables.

Run the same commands CI runs — [`.github/workflows/ci.yml`](.github/workflows/ci.yml)
is the source of truth if anything here drifts from it.

## Local hooks

`bun install` at the repo root (`prepare` script) installs
[lefthook](https://github.com/evilmartians/lefthook)'s git hooks; run
`bunx lefthook install` yourself if you skipped that step. What each hook
runs, straight from [`lefthook.yml`](lefthook.yml):

- **`pre-commit`** (staged files only, fixes and restages): `gofmt`,
  `go mod edit -fmt go.mod`, `biome check --write` for `web/`,
  `sort-package-json`, Prettier/`markdownlint` for Markdown, and a
  `docker build` when the CI `docker` group changed. The fixers take a
  file-type glob, since they act on the staged files themselves.
- **`commit-msg`**: [commitlint](https://commitlint.js.org) against
  `@commitlint/config-conventional`.
- **`pre-push`** (whole tree, never writes): `go vet`, `go test -race -cover`,
  `go mod tidy -diff`, `golangci-lint run` -- the Go commands run
  through Docker (`golang:<go.mod's version>-bookworm`,
  `golangci/golangci-lint:<CI's pin>`) so the version checking your push
  is always the one the repo declares, not whatever your package manager
  last updated -- plus `sort-package-json --check`, `bun run check`,
  `bun run lint`, `bun run test`, `bun run test:mutation`, `docker build`,
  `hadolint`, `goreleaser check`, Redocly, the `go.mod` formatting check
  and the Markdown/prose checks.

The Go hooks create `~/.cache/go-build-docker` and the other cache
directories themselves, so Docker doesn't create them as root on a clean
machine. lefthook prints nothing for a passing hook and the full output
for a failing one (`output: [failure]`).

Two CI checks have no hook, on purpose: `govulncheck` and `bun audit`.
Both read advisory databases that change without a commit, so a local pass
proves nothing the pipeline won't re-check. `lefthook.yml` says so too, and
`.github/scripts/lefthook-parity.test.sh` fails when any other check loses
its hook.

Each `pre-push` job runs through
[`hook-guard.sh`](.github/scripts/hook-guard.sh), which asks
[`changes.sh`](.github/scripts/changes.sh) whether that job's CI group
changed since `origin/main`. A hook and its CI job therefore run on the
same files, and a path added to `changes.sh` reaches both. To add a hook,
name its CI group as the guard's first argument; an unknown group fails the
hook instead of skipping it.

Vale runs through [`scripts/lint-prose.sh`](scripts/lint-prose.sh), which the
hooks and CI both call. It uses a `vale` on your `PATH` and otherwise runs the
official `jdkato/vale` image, pinned by tag and digest in the script, so
bumping Vale is one edit.

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
- CI has to be green (`build, vet, test`, `golangci-lint`, `redocly-lint`,
  `frontend`, `e2e`, `docker build` — all required checks on `main`) before a
  PR merges.
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

The release workflow's `screenshots` job then recaptures README.md's
screenshots against the new build and opens a PR with the diff for review.
That step needs the repo's Settings → Actions → General → "Allow GitHub
Actions to create and approve pull requests" enabled -- GitHub's default
`GITHUB_TOKEN` can't open a PR without it, and the job fails silently on
`gh pr create` otherwise.

## Which CI checks run

`ci.yml` runs a check only when files it covers change. Its first job,
`changes`, turns the changed files into one flag per kind of check, using the
paths in `.github/scripts/changes.sh`, and every other job runs on its flag.
A skipped job counts as a pass for the required checks. Editing `ci.yml` or
the script runs everything.

When a job starts reading a new kind of file, widen its group in
`changes.sh` and add a case to `.github/scripts/changes.test.sh` first. The
test runs in the `changes` job, so a missing path fails CI instead of
skipping a check that should have run. Run it locally with
`.github/scripts/changes.test.sh`.

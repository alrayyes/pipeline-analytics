# Proposal

## Why

Codecov shows the coverage and the test results, but not what a change does to
the size of the frontend bundle. A dependency that doubles the bundle goes
unnoticed until a Lighthouse run complains (#421).

## Research

- Codecov's bundle analysis for Vite is `@codecov/vite-plugin`. It reuses the
  upload token that coverage and test results use, so there's no new secret
  ([Codecov Tokens](https://docs.codecov.com/docs/codecov-tokens),
  [Vite quick start](https://docs.codecov.com/docs/vite-quick-start)).
- The plugin has a `dryRun` option. With it on, the build runs and nothing is
  uploaded. A build with a bogus token still finishes (checked locally), but
  a dry run without a token avoids three failed retries per build.
- The Go and frontend JUnit uploads the ticket also asks for are already in
  `ci.yml` and covered by the "Test results reach Codecov for both suites"
  requirement, so this change doesn't touch them.

## What changes

- `web/package.json` gains `@codecov/vite-plugin` at an exact version.
- `web/vite.config.ts` registers it last, with options from
  `web/src/bundleAnalysis.ts`: the `pipeline-analytics` bundle name, the token
  from `CODECOV_TOKEN`, telemetry off, and a dry run when there's no token.
- The `frontend` job's build step passes `CODECOV_TOKEN`.
- `README.md` and `CONTRIBUTING.md` say what the build uploads.

## Impact

- Modified capability: `ci-pipeline`, with one new requirement.

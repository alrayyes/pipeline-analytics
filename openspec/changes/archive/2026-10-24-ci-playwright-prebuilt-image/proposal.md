# Proposal

## Why

The e2e job and the README-screenshots job installed Bun and Playwright's
browsers on every run, about 20 s each time and a download that can hang or
flake (#735). A prebuilt image has them already, but only if its Playwright
version matches the `@playwright/test` the lockfile locks, and only if it can't
change under the workflow.

## What Changes

- Both jobs run in `ghcr.io/alrayyes/ci-playwright`, pinned by exact version and
  digest, with `--ipc=host` for Chromium.
- `.github/scripts/playwright-image.test.sh` fails the `changes` job and the
  pre-push hook when the image tag differs from the lockfile's version, when the
  digest is missing, when a job still runs `playwright install`, or when the
  container lacks `--ipc=host`.
- CONTRIBUTING.md describes the image and the check.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: a requirement that the CI Playwright image matches the
  lockfile and is pinned by digest.

## Impact

- `.github/workflows/ci.yml`, `.github/workflows/release.yml`,
  `.github/scripts/playwright-image.test.sh`, `lefthook.yml`, `CONTRIBUTING.md`.

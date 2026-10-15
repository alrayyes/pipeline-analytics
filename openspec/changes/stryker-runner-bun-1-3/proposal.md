# Proposal

## Why

`@hughescr/stryker-bun-runner` 1.5.1 and later need bun 1.4, and CI is pinned
to bun 1.3.14 because bun 1.4 writes a lockfile Dependabot can't read.
Dependabot's #613 took the runner to 1.5.2, so the mutation job fails on any
change under `web/` (#620). Main looks green only because nothing has changed
there since.

## What Changes

- `web/package.json` pins the runner to 1.5.0, the newest release that
  supports bun 1.3, and `bun.lock` follows.
- `.github/dependabot.yml` ignores runner releases from 1.5.1, with a note on
  when to lift it.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: a requirement that the dev tools CI runs work on the bun
  version CI pins.

## Impact

- `web/package.json`, `web/bun.lock`, `.github/dependabot.yml`.

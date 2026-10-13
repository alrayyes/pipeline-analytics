# Proposal

## Why

`bun audit` in CI carries `--ignore GHSA-hp3w-g68c-fv3c` for `sprintf-js`,
which has no patched release (#486). The only path to it is a dev tool:
`@lhci/cli` > `@lhci/utils` > `js-yaml` 3 > `argparse` 1 > `sprintf-js`. An
ignore that has to wait for a release that may never come hides the next
advisory on the same package.

## What Changes

- `web/package.json` overrides `js-yaml` to 4.3.2. `js-yaml` 4 depends on
  `argparse` 2, which has no dependencies, so `sprintf-js` leaves the lockfile.
- `@lhci/utils` only calls `js-yaml` (`safeLoad`) for a YAML Lighthouse config,
  and this repo's is `lighthouserc.cjs`. The full `web/scripts/lighthouse.sh`
  run passes with the override.
- The `--ignore` flag goes from CI and `CONTRIBUTING.md`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci-pipeline`: adds a requirement that the dependency audit runs without
  ignores.

## Impact

- `web/package.json`, `web/bun.lock`, `.github/workflows/ci.yml`,
  `CONTRIBUTING.md`.

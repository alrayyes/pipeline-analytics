# Proposal

## Why

The release history page asks GitHub's API for the releases from the
visitor's browser on every load. That request is unauthenticated, so it's
rate-limited per IP, adds a third-party round trip to each visit, and fails
whenever GitHub does. The content only changes when a release is cut, and
the notes are already in `CHANGELOG.md` at build time. Tracked in #368.

## What Changes

- The web build generates `releases.json` from `CHANGELOG.md` (release-please's
  own format) into the static assets, so it ships inside the binary.
- `/releases` loads that file instead of calling `api.github.com`.
- The "Couldn't load releases" state stays, for a missing or malformed file.
- The Playwright journey stops mocking GitHub's API and instead fails if the
  page requests it.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a release-history requirement: bundled with the
  build, no third-party request on load.

## Impact

- Frontend: `web/src/lib/changelog.ts` (parser), `web/scripts/generate-releases.ts`,
  `web/src/routes/releases/+page.svelte`, `web/package.json` scripts,
  `web/.gitignore` (the JSON is generated, not committed).
- Release: no workflow change. `.goreleaser.yml`'s hook already runs
  `bun run build`, which now includes the generation step, from the commit
  release-please tagged, so a release's own notes are in its own binary.
- The footer's version link and the GitHub releases link are unchanged.

# Proposal

## Why

The README showed nine screenshots and several pages had none. The capture
script matched the UI, but no spec said which pages the README covers, and
nothing stopped a new page being left out. Tracked in #549.

## What Changes

- One screenshot per page under `web/src/routes/`, except the footer's own
  pages (`/releases` and `/legal`), captured from mocked data.
- A unit test fails when a page has no screenshot, so a new page can't be
  forgotten.
- The README's screenshot section is generated from the same list between
  markers, by the capture script, so a new page needs no README edit.
- The release job's pull request carries the images and the README section,
  and now opens when the previous release's pull request has already merged.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `docs-screenshots`: adds the pages covered, the footer exclusion, the
  generated README section and the release pull request.

## Impact

- `web/scripts/capture-screenshots.mjs` and `.sh`, new
  `web/scripts/screenshot-routes.ts` and `update-readme-screenshots.ts`,
  `web/src/lib/screenshotRoutes.test.ts`.
- `.github/workflows/release.yml`, `README.md`, `CONTRIBUTING.md`, `docs/`.

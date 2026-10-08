# Proposal

## Why

The footer names the app and its version but doesn't say where the source is
or what license it's under. A visitor has to search for both. Tracked in #516.

## What Changes

- The footer links to the GitHub repository, with the GitHub mark beside the
  label.
- The footer names the license, AGPL-3.0, as a link to the `LICENSE` file.
- The links wrap on a narrow phone with no horizontal scroll, and no
  separator is left hanging at the end of a row.
- The privacy page says the two links are plain links and make no request on
  load.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement for the footer's source and license
  links.

## Impact

- `web/src/lib/components/Footer.svelte`, `web/src/routes/legal/+page.svelte`.
- `web/tests/e2e/footer.spec.ts`.
- Design: `docs/design/footer-screen.png` and `footer-links.png` (Stitch). The
  design separates items with middots, which leave one dangling where a row
  wraps, so the build spaces them instead.

# Proposal

## Why

The footer links the version to that release's page on GitHub and has a
separate "Release history" link beside it. The history page is bundled with
the app now (#370), so the version can open it directly and the second link
is redundant. Tracked in #384.

## What Changes

- The footer's version is a link to the release history page.
- The separate "Release history" link is removed.
- On the release history page itself the version is plain text, as the
  privacy link already is on its own page.
- A dev build still reads "dev build" with no link.
- The privacy page's description of the release history page is corrected: it
  said the page calls GitHub's API from the browser, which stopped being true
  with #370.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement for the footer's version link.

## Impact

- `web/src/lib/components/Footer.svelte`, `web/src/routes/legal/+page.svelte`.
- The release-history journey in `web/tests/e2e/dashboard.spec.ts` and a new
  `footer.spec.ts`.
- The link to a release's own page on GitHub moves off the footer: each card
  on the history page still links to it.

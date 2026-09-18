# Proposal

## Why

Theme and filter choices only survive in whichever browser set them:
`theme` and `forgeFilter` live in `localStorage`, and the Pipelines page's
health-status filter, repo selector, and sort order aren't persisted at
all — they reset on every navigation. None of it follows the user to a
different browser or device. That gap becomes a real problem once
`support-multiple-passkeys` lets the same account log in from more than
one device: settings need to live with the account, not the browser.

## What Changes

- New server-side settings record per account: theme (`light`/`dark`/
  `system`), forge filter, Pipelines health-status filter, Pipelines repo
  selector, and Pipelines sort order.
- New `GET`/`PUT` settings API, read on load and written on every change.
- The theme toggle moves off the always-visible `Nav.svelte` icon button
  and onto a new Settings page; the Nav button is removed.
- The Pipelines page's health filter, repo selector, and sort order stop
  being ephemeral component state and become part of the persisted
  settings, so they survive navigation and reload, not just across
  logins.
- The client keeps a local cache of the last-synced settings (replacing
  today's `localStorage`-as-source-of-truth role) purely to paint the
  right theme/filters instantly before the settings API responds; the
  server record is the source of truth and cross-device sync works
  through it.
- A "Reset filters" control on the Pipelines page resets its filter
  controls (health filter, repo selector, sort order) to their defaults
  and writes that back through the settings API; disabled (not hidden)
  when every filter already matches its default, matching this app's
  existing convention for a currently-inapplicable action (e.g. the
  Repos page's pagination buttons). This resets to defaults, not to an
  empty state — the filters are single-select toggle groups/selects,
  which don't have a meaningful "nothing selected" state.

## Capabilities

### New Capabilities

- `account-settings`: stores the dashboard account's UI preferences
  (theme, forge filter, Pipelines health filter, Pipelines repo
  selector, Pipelines sort order) server-side and serves them back on
  every login, from any device.

### Modified Capabilities

(none — no existing capability documents theme or filter behavior today)

## Impact

- Frontend: new `/settings` route; `web/src/lib/theme.svelte.ts` and
  `web/src/lib/forgeFilter.svelte.ts` change from `localStorage`
  read/write to syncing against the settings API (with `localStorage`
  demoted to a local paint cache); `web/src/lib/components/Nav.svelte`
  loses its theme toggle; the Pipelines page's health filter, repo
  selector, and sort order move out of local component state.
- Backend: new migration adding a settings table/row in `internal/db`; a
  new store and HTTP handler (`internal/httpserver`) for `GET`/`PUT
/api/settings`, gated by the existing session middleware.
- API contract: new `/api/settings` endpoints added to the OpenAPI spec
  under `openapi/`.

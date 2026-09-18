# Tasks

## 1. Backend: settings storage and API

- [x] 1.1 Add the `account_settings` migration (`user_id` PK referencing
      `webauthn_users`, `data TEXT`, `updated_at`) and verify it applies
      cleanly via the existing migration test harness
- [x] 1.2 Add a settings store (get/patch, with `null`-clears-to-default
      semantics) and unit tests covering: default returned when unset,
      update persists, `null` reverts to default, invalid value rejected
- [x] 1.3 Add `GET`/`PATCH /api/settings` handlers gated by the existing
      session middleware, and verify an unauthenticated request is
      denied (`401`), satisfying `dashboard-auth`'s session-gated-access
      requirement
- [x] 1.4 Add both endpoints to the OpenAPI spec under `openapi/` and
      verify it lints clean (Redocly)

## 2. Frontend: settings sync

- [x] 2.1 Rework `theme.svelte.ts` to sync through `GET`/`PATCH
    /api/settings`, keeping `localStorage` as a paint-only cache
      populated after each successful sync
- [x] 2.2 Rework `forgeFilter.svelte.ts` the same way
- [x] 2.3 Add persisted stores for the Pipelines page's health filter,
      repo selector, and sort order, replacing their component-local
      `$state`, and verify values survive a full page reload
- [x] 2.4 Wire the settings fetch into `+layout.ts`'s existing `load()`
      call, in parallel with the `/api/repos?limit=1` auth check, and
      verify no extra sequential round trip is added

## 3. UI: Settings page and filter reset

- [x] 3.1 Add a `/settings` route with the theme control (light/dark/
      system) and remove the theme toggle from `Nav.svelte`
- [x] 3.2 Add a "Reset filters" control to the Pipelines page that
      `PATCH`es the three filter keys to `null`, disabled when every
      filter already matches its default, and verify both the enabled
      and disabled states manually
- [x] 3.3 Run the Playwright suite (including the axe-core scan on the
      new Settings page) and verify it passes

## 4. Ship it

- [ ] 4.1 Open a pull request with `Closes #187` and verify CI passes

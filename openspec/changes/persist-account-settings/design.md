# Design

## Context

See `proposal.md` - Why. Relevant existing state:

- `web/src/lib/theme.svelte.ts` and `web/src/lib/forgeFilter.svelte.ts`
  are module-level Svelte 5 rune stores, each reading/writing its own
  `localStorage` key directly. The Pipelines page's health filter, repo
  selector, and sort order are component-local `$state`, never
  persisted.
- The app is a client-only SPA (`ssr: false`); there is no
  server-rendered HTML to seed from a cookie or otherwise, so
  `app.html`'s anti-flash script reads `localStorage` synchronously
  before hydration regardless of where the source of truth ends up
  living.
- The backend is a Go service (`internal/httpserver`, `internal/db`)
  behind a single-account WebAuthn session (`internal/auth`), backed by
  SQLite. `internal/db/migrations/00002_auth.sql` already stores
  WebAuthn credential and ceremony state as an opaque `data BLOB`
  column rather than one column per field.

## Goals / Non-Goals

**Goals:**

- One account-scoped settings record, readable and writable only by an
  authenticated session, that becomes the source of truth for theme
  and the persisted filters.
- No flash of the wrong theme/filters on load, despite the source of
  truth now requiring a network round trip.
- Adding a new persisted setting later doesn't require a schema
  migration.

**Non-Goals:**

- Multi-user settings (there is exactly one account; see
  `dashboard-auth`). No per-device or per-credential settings either -
  one record per account, full stop.
- Real-time sync between two concurrently open sessions (e.g. a
  WebSocket push when settings change elsewhere). A page reload or
  navigation is sufficient to pick up a change made on another device.

## Decisions

**Storage shape: one JSON blob column, not one column per setting.**
Mirrors the existing `webauthn_credentials`/`webauthn_ceremonies` data
columns rather than each WebAuthn field getting its own column. A new
migration:

```sql
CREATE TABLE account_settings (
    user_id    TEXT PRIMARY KEY REFERENCES webauthn_users (id) ON DELETE CASCADE,
    data       TEXT NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Only explicitly-set keys are stored in `data`; anything absent falls
back to a documented default resolved server-side at read time. This
means a new setting ships with a Go-side default constant and no
migration. Alternative considered: a fixed-columns table (one column
per setting, `NOT NULL DEFAULT`). Rejected because every new setting
(and this proposal already introduces five) would need its own
migration, and the JSON-blob precedent already exists in this codebase
for exactly this kind of "opaque, evolving structure" data.

**API: `GET`/`PATCH /api/settings`, not a resource per setting.**
`GET` returns every setting fully resolved (stored override merged
onto defaults - the response never has a null or missing field).
`PATCH` accepts a partial object; a key set to `null` clears that
key's stored override, reverting it to the documented default on the
next read. This makes "reset filters" a plain `PATCH` with the three
Pipelines filter keys set to `null` - no separate reset endpoint,
directly satisfying `account-settings`'s "single request" reset
requirement. The `null`-clears-to-default convention avoids hardcoding
default literals in both the client (to send them) and the server (to
resolve them); the client reset action doesn't need to know what the
defaults are.

**Client: `localStorage` demoted from source of truth to paint cache.**
`theme.svelte.ts` and `forgeFilter.svelte.ts` (plus new stores for the
three Pipelines filters) keep writing to `localStorage` on every
change, but now as a side effect after a successful `PATCH`, not as
the persistence mechanism itself. `initTheme()`/`initForgeFilter()`
(and the new filter initializers) read the cached `localStorage` value
first for an instant, correct-looking paint, then fetch `GET
/api/settings` and reconcile - overwriting local state (and the cache)
if the server disagrees, e.g. because a different device changed it
since the cache was last written. `app.html`'s inline anti-flash
script is unchanged: it already reads `localStorage['theme']`
synchronously, which is exactly the cache this design keeps populated.

**Settings fetch folds into the existing auth-gate request.**
`web/src/routes/+layout.ts`'s `load()` already calls `/api/repos?limit=1`
on every navigation to detect a `401` and redirect to `/login`. Settings
fetch is a second call from the same `load()`, run in parallel, so it
does not add a new round trip to the navigation-gating path.

## Risks / Trade-offs

- **Stale cache after cross-device change** -> the `localStorage` paint
  cache can briefly show an outdated value on load before the `GET`
  response reconciles it. Mitigated by resolving before the first
  meaningful render blocks on it only for the anti-flash theme case
  (already true today); filter UI reconciles on mount same as any
  other async load.
- **Unbounded JSON blob growth** -> a new setting added later has
  nothing stopping it from bloating `data` indefinitely. Mitigated by
  code review discipline (single account, low field count expected)
  rather than a schema constraint; revisit if this ever becomes
  multi-user.

## Migration Plan

1. Add the `account_settings` migration.
2. Add the Go store + `GET`/`PATCH /api/settings` handlers, gated by
   the existing session middleware.
3. Add the OpenAPI definition for both endpoints.
4. Migrate `theme.svelte.ts` and `forgeFilter.svelte.ts` to sync
   through the API (cache-then-reconcile), and add equivalent stores
   for the three Pipelines filters.
5. Move the theme control from `Nav.svelte` to a new Settings page;
   add the "Reset filters" control to the Pipelines page.

No rollback concern beyond the reverse migration: nothing downstream
depends on `account_settings` existing, and the client falls back to
its `localStorage` cache/defaults if the API is temporarily
unreachable.

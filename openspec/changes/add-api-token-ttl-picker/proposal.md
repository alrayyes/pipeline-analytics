# Proposal

## Why

`add-api-token-auth` (#178, PR #186) ships `POST /api/auth/tokens` with a
hardcoded fixed 1-year TTL — a deliberate call at the time (see that
change's `design.md`: "a long, fixed TTL rather than no expiry at all"),
but Ryan now wants to pick the lifetime at creation instead: a
short-lived CI credential shouldn't carry the same year-long blast
radius as one meant to stick around. Tracked as
[alrayyes/pipeline-analytics#191](https://github.com/alrayyes/pipeline-analytics/issues/191).
PR #186 is still open/unmerged as of this proposal, so this lands as a
follow-up on the same capability rather than reopening #178 — see
design.md's "Sequencing" for what that means for this change's own
delta spec.

There's also no frontend for token management at all yet — #186 is
backend-only (confirmed: `git show origin/add-api-token-auth --stat`
touches no `web/` files) — so this change has to build the minimal UI
needed to host the picker, not just the picker itself.

## What Changes

- `POST /api/auth/tokens` accepts an optional requested TTL; the server
  clamps it to a documented hard ceiling regardless of what the client
  sends — the ceiling is the enforcement boundary, the client-offered
  presets are UX only.
- New token-management page: lists nothing (no listing endpoint exists
  or is being added — out of scope, same as #178's own non-goals) but
  hosts token creation, a preset TTL picker built on the existing
  `ToggleGroup` pattern (`web/src/lib/components/ForgeFilter.svelte`),
  and a live absolute-expiry preview that updates as the selection
  changes.
- No "no expiration" option anywhere in the picker — every token still
  expires, unchanged from #178's constraint.
- `openapi/openapi.yaml`'s token-creation request schema gains the new
  optional TTL field.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-auth`: the "API token issuance and revocation" requirement
  (added by the still-unmerged `add-api-token-auth`) gains a
  client-selectable TTL, server-clamped to a ceiling, plus the UI to
  drive it. See design.md's "Sequencing" for how this delta spec relates
  to that one.

## Impact

- `internal/auth/auth.go` / `internal/auth/service.go`: issuance takes
  an optional requested TTL, clamps it server-side.
- `internal/httpserver/auth.go`: token-creation handler reads the new
  request field.
- `openapi/openapi.yaml`: new optional field on the token-creation
  request body.
- Frontend: new route for token creation (own page, not nested under
  the not-yet-built `/settings` route from `persist-account-settings` —
  see design.md's "Alternatives considered"); new component built on
  `web/src/lib/components/ui/toggle-group/`.
- Accessibility: the new picker needs the same `role="radiogroup"`
  wrapper `ForgeFilter.svelte` already uses (bits-ui's `ToggleGroup`
  root renders `role="group"`, and `type="single"` items render
  `role="radio"`, which axe's `aria-required-parent` flags without an
  explicit `radiogroup` ancestor) and an axe-core scan per this repo's
  accessibility-testing rule.

# Proposal

## Why

Only one WebAuthn credential can ever be registered: `BeginRegistration`
rejects any registration attempt once an account exists
(`internal/auth/service.go`), so the dashboard is unusable from a second
device (phone, a second laptop) without sharing the same passkey across
them, which most authenticators don't support. `persist-account-settings`
also makes cross-device use of the account a real workflow, not just a
login formality — settings should follow you between devices, which only
matters if you can actually log in from more than one.

## What Changes

- A new, authenticated "add a passkey" ceremony lets the already-logged-in
  account register an additional WebAuthn credential, separate from the
  existing anonymous first-run registration (which still creates exactly
  one _account_ and still rejects a second one).
- A Settings page section lists every registered passkey (by a
  user-supplied label and creation date), lets the user register another,
  and lets them revoke one they no longer use.
- **BREAKING**: none. The existing first-run/single-account behavior is
  unchanged; this only adds a second, authenticated enrollment path.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-auth`: adds authenticated credential enrollment and
  credential management (list, label, revoke) alongside the existing
  first-run registration and login requirements, which are unchanged.

## Impact

- Backend: `internal/auth` gains an authenticated "begin/finish add
  credential" ceremony (reusing `webauthn.BeginRegistration`/
  `FinishRegistration` against the existing user instead of a temp user),
  plus list/label/revoke operations on `webauthn_credentials`.
  `internal/db`: a new column (or small table) for a credential's
  user-supplied label. `internal/httpserver`: new authenticated endpoints
  for add/list/label/revoke.
- Frontend: a Settings page section for credential management; the
  existing `web/src/routes/login/+page.svelte` registration flow is
  unchanged (it only ever handles the first, anonymous registration).
- API contract: new endpoints added to `openapi/`.

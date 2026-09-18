# Design

## Context

See `proposal.md` - Why. Relevant existing state:

- `internal/auth/service.go`'s `BeginRegistration` hard-rejects if
  `store.GetUser` finds an existing user at all - there is no
  authenticated path into a registration ceremony today, only the
  anonymous first-run one.
- The schema already models one-to-many correctly:
  `webauthn_credentials.user_id` references `webauthn_users.id`
  (`internal/db/migrations/00002_auth.sql`), and `PutCredential` does
  `INSERT ... ON CONFLICT (id) DO UPDATE` keyed by the credential's own
  `id` (`internal/auth/sqlite/store.go`) - a genuinely new credential
  inserts a new row rather than overwriting. Nothing about storage
  needs to change to hold more than one credential per account.
- `internal/httpserver/session.go`'s `requireSession()` middleware
  already gates everything except a documented allowlist
  (`isPublicPath()`), so an authenticated "add credential" endpoint
  gets that gating for free by simply not being on the allowlist.

## Goals / Non-Goals

**Goals:**

- Register a second (third, ...) passkey without touching the existing
  anonymous first-run/single-account registration path at all.
- Never leave the account with zero usable credentials.

**Non-Goals:**

- Multi-user accounts. This is still exactly one account
  (`dashboard-auth`'s existing scope); it can just now log in with more
  than one credential.
- Renaming/re-labeling a credential after creation. Out of scope for
  this change - label is set once at enrollment; revoke-and-re-add is
  the escape hatch if a label needs to change.

## Decisions

**Reuse `webauthn.BeginRegistration`/`FinishRegistration` against the
real user, not a temp one.** `BeginRegistration`'s existing anonymous
path constructs a `tempUser` with a fresh random handle because no
account exists yet. The new authenticated path instead loads the real
`User` via `store.GetUser` (already used elsewhere via
`userWithCredentials`) and passes their existing `UserHandle` and
credential list to `webauthn.BeginRegistration`, so the ceremony
correctly excludes credentials already registered
(`ExcludeCredentials`, standard WebAuthn behavior for "register
another"). This is almost entirely a new `Service` method
(`BeginAddCredential`/`FinishAddCredential`) sharing `saveCeremony`/
`loadCeremony`, not new library-level logic.

**Label stored as a plain column, not derived from authenticator
data.** WebAuthn doesn't reliably expose a human-friendly authenticator
name across platforms/browsers, so a new `label TEXT NOT NULL`
column on `webauthn_credentials` is set from a value the client
collects at enrollment time (a text field: "MacBook", "iPhone"). No
new table - it's one credential-scoped fact, not a separate entity.
Alternative considered: derive a name from AAGUID-to-authenticator-name
lookup tables. Rejected - those lists go stale and still need a
manual-override path, so the manual path might as well be the only
path.

**Credential identifier over the wire: base64url of the credential
ID.** `webauthn_credentials.id` is a `BLOB` (raw credential ID); the
list/revoke API needs a JSON-safe identifier, so responses/requests use
the standard base64url encoding already used for other WebAuthn binary
fields in this codebase's ceremony JSON.

**Revoke checks `CredentialsForUser` count, not session state.** "Last
remaining credential" is evaluated as "does this account have more than
one credential row right now", independent of which credential
authenticated the current session - simpler than tracking a
per-session credential identity, and matches the spec's actual
wording.

## Risks / Trade-offs

- **A revoked credential's still-active session isn't invalidated** ->
  a session already established stays valid until its normal
  expiry even after its credential is revoked; only future logins
  with that credential are blocked. Mitigated by the existing session
  TTL already bounding exposure; treated as acceptable for a
  single-user dashboard, not a multi-tenant system.
- **Concurrent revoke of the last two credentials** -> a race between
  two requests each seeing "2 credentials, safe to revoke" could zero
  the account out. Mitigated by making the count-check and delete a
  single transaction in the store, not a check-then-act across two
  round trips.

## Migration Plan

1. Add a migration for `webauthn_credentials.label TEXT NOT NULL
DEFAULT ''` (existing rows get an empty label, editable only via
   revoke-and-re-add per the stated non-goal).
2. Add `BeginAddCredential`/`FinishAddCredential` to `internal/auth`
   plus list/revoke store methods (transactional revoke per the
   trade-off above).
3. Add the authenticated HTTP endpoints in `internal/httpserver` and
   their OpenAPI definitions.
4. Add the Settings-page credential list/add/revoke UI.

No rollback concern beyond the reverse migration: existing single-
credential accounts are unaffected until they add a second passkey.

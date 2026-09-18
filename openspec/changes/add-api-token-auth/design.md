# Design

## Context

`internal/auth` already has a working pattern for an opaque, random,
server-side-verified credential: `sessions` (`internal/auth/sqlite/store.go`),
a 32-byte hex ID (`crypto.RandomHex`) stored **as plaintext** in SQLite and
looked up by exact match, with an in-DB `expires_at`. `requireSession`
(`internal/httpserver/session.go`) reads it off the `session` cookie for
every path `isPublicPath` doesn't exempt. `/api/auth/*` is currently
exempt wholesale — register/login/logout have to be reachable
unauthenticated, but a token-management endpoint under the same prefix
would inherit that exemption by accident if `isPublicPath` isn't taught
the difference.

See proposal.md for the full shape; this covers how.

## Goals / Non-Goals

**Goals:**

- Reuse the session pattern's shape (`Store` port method, SQLite
  adapter, goose migration) so a token isn't a second parallel
  authentication system, just a second credential kind the same gate
  accepts.
- Keep the raw token unrecoverable after creation.

**Non-Goals:**

- Token scopes/permissions narrower than the single dashboard user
  already has — this app has exactly one account (`dashboard-auth`'s
  own "WebAuthn credential registration" requirement), so a token is
  either valid for that account or it isn't.
- A token-listing endpoint. Neither #178's acceptance criteria nor its
  Definition of Done ask for one, and a revoke-by-id flow doesn't need
  it — add one later if a real need shows up.
- Updating the four SDK repos from #155 — they don't exist yet.

## Decisions

- **Hash the token at rest (SHA-256), unlike the session ID's
  plaintext storage.** A session ID lives only in an `HttpOnly` cookie
  with a short-ish practical exposure window (the browser, one
  origin); a token is designed to be copied into a script, an env
  var, a CI secret — more places it can leak from, so it shouldn't
  also be recoverable by reading the database. The raw value is
  returned exactly once, at creation, the same tradeoff most API
  providers make (GitHub PATs, Stripe keys). SHA-256 (not bcrypt/argon2)
  because the token itself is already a 32-byte random value, not a
  human-guessable password — there's no low-entropy input to slow
  down brute-forcing.
- **Token management endpoints are session-only, not bearer-eligible,**
  even though every other gated endpoint now accepts either. A token
  that can mint or revoke other tokens turns one leaked credential
  into an unbounded blast radius; requiring the browser session for
  that specific pair of endpoints caps a leaked token at "whatever the
  API itself exposes," matching the threat model implied by #178's
  own framing (a token is for headless *use*, not account
  administration).
- **`isPublicPath` gets a token-path exception rather than moving the
  new endpoints outside `/api/auth/`.** Keeping them under the
  existing prefix matches where `register`/`login`/`logout` already
  live and keeps the OpenAPI path grouping coherent; the alternative
  (a new top-level prefix) buys nothing and splits auth-related paths
  across two places in the spec.
- **A long, fixed TTL (1 year) rather than no expiry at all.** #178's
  acceptance criteria explicitly test "expired or revoked," so a token
  has to be able to expire on its own. A year is long enough that a
  script relying on it won't need to babysit rotation, short enough
  that a token nobody remembers issuing doesn't stay valid forever.
  Revocation is what actually terminates casual use; the TTL is the
  backstop for a token everyone's forgotten about.

## Risks / Trade-offs

- A hashed, unrecoverable token means a user who loses it has to issue
  a new one and revoke the old — no "show me my token again" recovery
  path. Same trade every hashed-secret API makes; documenting it in
  the endpoint's response is enough, no extra mechanism needed.
- Two credential kinds now flow through one gate (`requireSession`,
  kept as the name since it's still the same call site in
  `server.go` — the doc comment explains what it now also accepts).
  Read carefully once; a future auth change touching that function
  needs to remember both paths exist.

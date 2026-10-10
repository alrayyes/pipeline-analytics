# Design

## Context

`internal/auth/sqlite/store.go` hardcodes the lifetime:

```go
const apiTokenTTL = 365 * 24 * time.Hour
```

`Store.CreateToken` takes no lifetime and sets `ExpiresAt: now.Add(apiTokenTTL)`.
`Service.IssueToken(ctx, userID)` and the HTTP handler `issueToken`
(`internal/httpserver/auth.go`) pass nothing through. No page creates
tokens; the Settings page exists but has no token section. See proposal.md
for why.

## Goals / Non-Goals

**Goals:**

- A requested lifetime flows from the request through `Service` to `Store`,
  clamped at the layer nearest persistence, so no caller can bypass the
  ceiling by calling `Service` directly.
- One constant plays the ceiling's role, so there is no second "max" to
  drift from it.

**Non-Goals:**

- Listing or revoking tokens in the UI. The revoke endpoint exists; a list
  endpoint and the UI for both are a separate ticket if wanted.
- Changing how tokens authenticate requests, or the session-only gating on
  issuing and revoking.
- A floor on the lifetime. The clamp is a ceiling; a one-second token only
  hurts its owner.

## Decisions

- **Replace `apiTokenTTL` with `auth.TokenTTLCeiling` (365 days) and
  `auth.TokenTTLDefault` (90 days)**, exported from the `auth` package where
  the store, the handler and the tests can all name them. The old name reads
  as "the TTL", which stops being true once a request can come in under it.
  Ninety days is also the picker's default preset: one value, not two to
  keep in sync.
- **`Store.CreateToken` takes `requestedTTL time.Duration` and clamps it
  itself** rather than `Service` passing a final `expiresAt`. The ceiling
  keeps one place it is applied, and an unset (zero) request falls back to
  the default in the same spot. `Service.IssueToken` becomes a pass-through.
- **The request field is `ttlSeconds` (optional integer), not an enum of
  preset names.** The server needs a duration to clamp, not the concept of
  "30 days". Zero or negative is a `400`; everything else clamps. A caller
  that sends no body, or no field, gets the default, so existing `curl`
  usage keeps working.
- **The response carries `expiresAt`**, so a client shows the expiry without
  a second call.
- **The picker is a section of the Settings page**, not its own route. The
  page now exists and already holds account settings, which is where someone
  looks for this.
- **The expiry preview is computed in the browser** (`now + preset`); the
  ceiling and default are static, so there is nothing to ask the server. The
  server's clamp stays authoritative.

## Alternatives considered

- **A standalone `/tokens` page** (this change's first version). It was
  chosen when no Settings page existed. Now it would be a second place for
  account configuration.
- **An enum of preset strings in the request.** It makes the API contract
  decide which presets exist, so a fifth preset would be a backend change.
- **Keeping 365 days as the default.** Rejected: the point of the ticket is
  that the short lifetime should be the easy one. It changes behaviour for a
  caller that sends no body; see Risks.

## Risks / Trade-offs

- **The default shortens from 365 to 90 days.** A script that created a
  token without a body and expected a year will find it expires after 90
  days. The README says how to ask for longer. There is no way to know which
  tokens are in use from the server, so the change is announced in the
  release notes via its commit type.
- Renaming `apiTokenTTL` touches every reference (`CreateToken`, its doc
  comment, and the tests that assert on it); all change in one commit.

## Migration Plan

No data migration. `api_tokens.expires_at` already stores an absolute
timestamp; only the value computed before the `INSERT` changes. Existing
tokens keep their expiry.

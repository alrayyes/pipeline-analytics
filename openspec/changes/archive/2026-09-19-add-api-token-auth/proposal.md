# Proposal

## Why

The only auth scheme is `sessionCookie`, established through the WebAuthn
ceremony — fine for the dashboard itself, but it leaves every generated
client SDK's "authentication" story as "open a browser, log in, copy the
raw session cookie into an environment variable," with no visible expiry
the SDK can react to. Tracked as
[alrayyes/pipeline-analytics#178](https://github.com/alrayyes/pipeline-analytics/issues/178),
surfaced while building the Go SDK.

## What Changes

- New `POST /api/auth/tokens` endpoint (session-gated) that issues a
  long-lived, revocable API token, returned once at creation.
- New `DELETE /api/auth/tokens/{tokenId}` endpoint (session-gated) that
  revokes a token.
- `requireSession` middleware accepts `Authorization: Bearer <token>` as
  an alternative to the session cookie for every session-gated endpoint
  except token issuance/revocation themselves, which stay session-only
  (a leaked token can't be used to mint or revoke tokens of its own —
  see design.md).
- A revoked or expired token gets `401` with the standard `Error` shape,
  same as an invalid session.
- `openapi/openapi.yaml` gains a second `securitySchemes` entry
  (`apiKey`/`http bearer`) alongside `sessionCookie` — additive, not a
  replacement; the dashboard itself keeps using cookies.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-auth`: "Session-gated access" now accepts a valid bearer
  token as an alternative to a session cookie; new requirement for
  token issuance and revocation.

## Impact

- `internal/auth/auth.go` (`Store` interface, new `Token` type, new
  sentinel errors).
- `internal/auth/sqlite/store.go` (+ new migration
  `internal/db/migrations/0000X_api_tokens.sql`).
- `internal/auth/service.go` (issuance/revocation/authentication logic).
- `internal/httpserver/session.go` (`requireSession` accepts bearer
  tokens; `isPublicPath` keeps token endpoints session-gated).
- `internal/httpserver/auth.go` (two new handlers).
- `openapi/openapi.yaml` (new scheme, two new paths, request/response
  schemas).
- No SDK repo changes here — #155's four repos don't exist yet; their
  READMEs/constructors pick this up once they do, per #178's own
  Definition of Done.

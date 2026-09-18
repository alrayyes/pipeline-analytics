# Tasks

## 1. Domain and storage

- [x] 1.1 Write a failing `sqlite/store_test.go` case: `CreateToken`
      returns a raw token once, `TokenUserID` resolves it back to the
      user, and the token is not recoverable from the row itself
- [x] 1.2 Add migration `internal/db/migrations/00004_api_tokens.sql`
      (goose up/down) for an `api_tokens` table: id, user_id,
      token_hash, expires_at, revoked_at nullable, created_at; verify
      `go test ./internal/db/...` still passes
- [x] 1.3 Add `Token` type, `ErrTokenNotFound`, and
      `CreateToken`/`TokenUserID`/`RevokeToken` to the `auth.Store`
      interface (`internal/auth/auth.go`)
- [x] 1.4 Implement the three methods on the SQLite adapter
      (`internal/auth/sqlite/store.go`): SHA-256 the raw token before
      storing/querying, 1-year TTL, `RevokeToken` scoped to `userID`;
      verify task 1.1's test now passes

## 2. Service layer

- [x] 2.1 Write a failing `service_test.go` case: `IssueToken` returns
      a raw value that `AuthenticateToken` resolves, a revoked token's
      raw value no longer resolves
- [x] 2.2 Add `IssueToken`, `RevokeToken`, `AuthenticateToken` to
      `internal/auth/service.go`, mapping store errors to the
      package's sentinel errors; verify 2.1 passes

## 3. HTTP layer

- [x] 3.1 Write a failing `auth_test.go` case: `POST /api/auth/tokens`
      without a session is `401`; with a session, `201` with a raw
      token in the body
- [x] 3.2 Write a failing case: a request to any session-gated
      endpoint (e.g. `GET /api/repos`) carrying `Authorization: Bearer
      <token>` for a token from 3.1 succeeds without a session cookie
- [x] 3.3 Write a failing case: the same request after `DELETE
      /api/auth/tokens/{tokenId}` (with a session) gets `401` with the
      standard `Error` body
- [x] 3.4 Write a failing case: `POST /api/auth/tokens` and `DELETE
      /api/auth/tokens/{tokenId}` called with only a bearer token (no
      session) are both `401`
- [x] 3.5 Update `isPublicPath` (`internal/httpserver/session.go`) so
      `/api/auth/tokens` and `/api/auth/tokens/{id}` stay gated while
      `/api/auth/register*`, `/login*`, `/logout` stay public
- [x] 3.6 Update `requireSession` to accept a valid bearer token as an
      alternative to the session cookie for every path except the
      token endpoints themselves; verify all of section 3's tests pass
- [x] 3.7 Add `issueToken`/`revokeToken` handlers to
      `internal/httpserver/auth.go` and wire both routes in
      `server.go`

## 4. OpenAPI spec

- [x] 4.1 Add an `apiKey`/`http bearer` entry to `securitySchemes` in
      `openapi/openapi.yaml`, additive alongside `sessionCookie`
- [x] 4.2 Add `POST /api/auth/tokens` and `DELETE
      /api/auth/tokens/{tokenId}` paths, request/response schemas, and
      `401` `Error` responses, scoped to `sessionCookie` security only
- [x] 4.3 Run `bunx @redocly/cli lint openapi/openapi.yaml` (per
      `rules/api.md`) and verify it's clean

## 5. Verify and ship

- [x] 5.1 Run `go test ./...` and verify everything passes
- [x] 5.2 Run `golangci-lint run` and verify it's clean
- [x] 5.3 Open a pull request with `Closes #178` and verify CI passes

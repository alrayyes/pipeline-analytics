# Tasks

## 1. Storage layer

- [x] 1.1 Write a failing `store_test.go` case: `CreateToken` with a
      requested lifetime under the ceiling issues a token expiring at
      `now + requested`
- [x] 1.2 Write a failing case: a requested lifetime above the ceiling
      clamps to `now + apiTokenTTLCeiling`
- [x] 1.3 Write a failing case: a zero requested lifetime issues a token
      expiring at `now + apiTokenTTLDefault`
- [x] 1.4 Replace `apiTokenTTL` with `auth.TokenTTLCeiling` and
      `auth.TokenTTLDefault = 90 * 24 * time.Hour`, and give `CreateToken` a
      `requestedTTL time.Duration`, clamping per design.md; 1.1 to 1.3 pass
- [x] 1.5 Update the `auth.Store` interface (`internal/auth/auth.go`)

## 2. Service layer

- [x] 2.1 Write a failing `service_test.go` case: `IssueToken` passes a
      requested lifetime through to the store unchanged
- [x] 2.2 Update `Service.IssueToken` to accept and pass it; 2.1 passes

## 3. HTTP layer

- [x] 3.1 Write a failing `auth_test.go` case: `POST /api/auth/tokens` with
      `{"ttlSeconds": <under the ceiling>}` issues a token whose response
      `expiresAt` matches
- [x] 3.2 Write a failing case: `ttlSeconds` above the ceiling clamps in the
      response, not just in storage
- [x] 3.3 Write a failing case: `ttlSeconds` of zero or negative is `400`
      and issues no token
- [x] 3.4 Write a failing case: no body, and a body with no `ttlSeconds`,
      still succeed with the 90-day default
- [x] 3.5 Update the `issueToken` handler to read the optional `ttlSeconds`
      and return `expiresAt`; 3.1 to 3.4 pass

## 4. OpenAPI and docs

- [x] 4.1 Add optional `ttlSeconds` (integer) to the `POST /api/auth/tokens`
      request and `expiresAt` to its response in `openapi/openapi.yaml`,
      with the default and the ceiling in the descriptions
- [x] 4.2 Run Spectral and Redocly; both pass
- [x] 4.3 Update the README's token instructions: the default is 90 days and
      `ttlSeconds` asks for longer, up to 365

## 5. Frontend

- [ ] 5.1 The API tokens section on Settings (presets, expiry preview, token
      shown once, axe scan, e2e journey) is its own ticket, tracked in
      #562; it builds on this change's API

## 6. Verify and ship

- [x] 6.1 `go test -race ./...` passes
- [x] 6.2 `golangci-lint run` is clean
- [ ] 6.3 Open the pull request (the API half of #191; the issue closes with #562's pull request) and verify CI passes
- [ ] 6.4 Archive once 1 to 5 have shipped

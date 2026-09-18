# Tasks

## 1. Storage layer

- [ ] 1.1 Write a failing `store_test.go` case: `CreateToken` with a
      requested TTL under the ceiling issues a token expiring at
      `now + requested`, not `now + ceiling`
- [ ] 1.2 Write a failing case: `CreateToken` with a requested TTL
      above the ceiling clamps to `now + apiTokenTTLCeiling`
- [ ] 1.3 Write a failing case: `CreateToken` with a zero/unset
      requested TTL issues a token expiring at `now +
      apiTokenTTLDefault`
- [ ] 1.4 Rename `apiTokenTTL` to `apiTokenTTLCeiling`, add
      `apiTokenTTLDefault = 90 * 24 * time.Hour`
      (`internal/auth/sqlite/store.go`), and change `CreateToken`'s
      signature to accept `requestedTTL time.Duration`, clamping it per
      design.md's "Decisions"; verify 1.1-1.3 pass
- [ ] 1.5 Update `auth.Store`'s `CreateToken` interface signature
      (`internal/auth/auth.go`) to match

## 2. Service layer

- [ ] 2.1 Write a failing `service_test.go` case: `IssueToken` passes
      a requested TTL through to the store unchanged (mock/fake store
      asserting the argument it received)
- [ ] 2.2 Update `Service.IssueToken` (`internal/auth/service.go`) to
      accept and pass through `requestedTTL time.Duration`; verify 2.1
      passes

## 3. HTTP layer

- [ ] 3.1 Write a failing `auth_test.go` case: `POST /api/auth/tokens`
      with `{"ttlSeconds": <value under the ceiling>}` in the body
      issues a token whose `expiresAt` in the response matches
- [ ] 3.2 Write a failing case: `POST /api/auth/tokens` with
      `ttlSeconds` above the ceiling (in seconds) clamps in the
      response, not just in storage
- [ ] 3.3 Write a failing case: `POST /api/auth/tokens` with
      `ttlSeconds: 0` or negative is rejected `400`
- [ ] 3.4 Write a failing case: `POST /api/auth/tokens` with no body
      (or no `ttlSeconds`) still succeeds, using the default
- [ ] 3.5 Update `issueToken` handler (`internal/httpserver/auth.go`)
      to parse an optional `ttlSeconds` field from the request body and
      pass it through; verify 3.1-3.4 pass
- [ ] 3.6 Add `expiresAt` to the token-creation response DTO if not
      already present, so the frontend can render it without a second
      call

## 4. OpenAPI spec

- [ ] 4.1 Add optional `ttlSeconds` (integer) to the `POST
      /api/auth/tokens` request schema and `expiresAt` to its response
      schema in `openapi/openapi.yaml`, documenting the default and
      ceiling in the field description
- [ ] 4.2 Run `bunx @redocly/cli lint openapi/openapi.yaml` (per
      `rules/api.md`) and verify it's clean

## 5. Frontend: token creation page

- [ ] 5.1 Add a `TtlPicker` component built on
      `web/src/lib/components/ui/toggle-group/` (same shape as
      `ForgeFilter.svelte`, including its `role="radiogroup"` wrapper),
      offering 30-day/90-day/1-year presets with 90 days selected by
      default
- [ ] 5.2 Wire the picker's selection to a live absolute-expiry preview
      (`now + selected preset`, formatted per this app's existing date
      display convention)
- [ ] 5.3 Add the `/tokens` route: a page that creates a token via
      `POST /api/auth/tokens` with the selected preset's TTL in
      seconds, and displays the raw token value once on success (per
      #178's existing "returned once" behavior)
- [ ] 5.4 Add a Playwright journey test for the page covering: default
      preset selected, expiry preview updates on selection change,
      token creation succeeds and displays the raw value once
- [ ] 5.5 Add an axe-core scan to that same journey test per this
      repo's accessibility-testing rule, and verify it's clean
      (particularly the `radiogroup` ancestor for the toggle group's
      `role="radio"` items)

## 6. Verify and ship

- [ ] 6.1 Run `go test ./...` and verify everything passes
- [ ] 6.2 Run `golangci-lint run` and verify it's clean
- [ ] 6.3 Run the frontend test suite and verify it's clean
- [ ] 6.4 Open a pull request with `Closes #191` and verify CI passes

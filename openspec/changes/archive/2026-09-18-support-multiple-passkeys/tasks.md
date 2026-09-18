# Tasks

## 1. Backend: enrollment, listing, revocation

- [x] 1.1 Add the `webauthn_credentials.label` migration and verify it
      applies cleanly via the existing migration test harness
- [x] 1.2 Add `Service.BeginAddCredential`/`FinishAddCredential`,
      loading the real user (not a temp one) and excluding already-
      registered credentials, with unit tests covering a second
      credential enrolling successfully against the same account
- [x] 1.3 Add a transactional revoke store method that rejects
      revoking an account's last remaining credential, with a unit
      test for both the normal and last-credential-rejected cases
- [x] 1.4 Add a credential-listing store method returning label and
      creation date per credential, with a unit test for more than one
      credential
- [x] 1.5 Add authenticated `POST` (begin/finish add), `GET` (list),
      and `DELETE` (revoke) endpoints in `internal/httpserver`, and
      verify each is denied without a valid session
- [x] 1.6 Add the new endpoints to the OpenAPI spec under `openapi/`
      and verify it lints clean (Redocly)

## 2. Frontend: credential management UI

- [x] 2.1 Add a credential-management section to the Settings page:
      list with label/date, an "Add a passkey" flow prompting for a
      label, and a revoke action per credential
- [x] 2.2 Disable/hide the revoke action on the only remaining
      credential, matching the backend's last-credential guard, and
      verify manually with one and with two credentials registered
- [x] 2.3 Run the Playwright suite (including the axe-core scan on the
      updated Settings page) and verify it passes

## 3. Ship it

- [x] 3.1 Open a pull request with `Closes #188` and verify CI passes

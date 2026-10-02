# Proposal

## Why

The settings page disables a passkey's Revoke button when the account has
one credential. That is a copy of a rule the server already enforces (a
revoke of the last credential is a 409), kept in the browser so the button
looks right. If the rule changed, the two would drift, and a client other
than the dashboard has to learn it by trying. Finding 3 of #376: derived
permissions come from the API, and hiding a button is cosmetic.

## What Changes

- `GET /api/auth/credentials` reports `revocable` on each credential, true
  unless it is the account's last remaining one, computed beside the
  server's own last-credential rule.
- The settings page disables Revoke from `revocable` and no longer counts
  credentials.
- The server's 409 stays: it is the enforcement, the flag is the hint.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-auth`: adds a requirement that the credential list says which
  credentials can be revoked.

## Impact

- `openapi/openapi.yaml` (`revocable` on `Credential`), `internal/auth`,
  `internal/httpserver`.
- `web/src/routes/settings/+page.svelte`. The existing e2e journey, which
  adds a second passkey and checks Revoke enables and disables, is the
  end-to-end acceptance test and is unchanged.

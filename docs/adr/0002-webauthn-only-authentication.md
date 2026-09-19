# 2. WebAuthn-only authentication, no password path

## Status

Accepted

## Context

The dashboard supports exactly one account and is reachable at a public
HTTPS URL (required for GitHub/Forgejo webhook delivery). It needs
login, but has no multi-user requirement and no team to administer
password resets for.

## Decision

Login is WebAuthn (passkey) only. There is no password path and no
forgotten-password flow. First run registers a passkey; every later
visit authenticates with one. Additional passkeys can be enrolled once
logged in (see `openspec/changes/archive/2026-09-18-support-multiple-passkeys/`).

## Consequences

- No credential-phishing surface: there's no password to phish, and a
  passkey is bound to this origin.
- No password-reset support burden -- there's exactly one account, held
  by the person who deployed the tool, so a forgotten password has no
  one else to serve it to.
- Losing every enrolled authenticator with no backup is a real lockout
  risk with no recovery flow. Mitigated by supporting more than one
  enrolled passkey per account (2026-09-18) rather than by adding a
  password fallback, which would reopen the phishing surface this
  decision exists to avoid.
- This is a single-tenant decision. Multi-user support would need this
  revisited, not just extended -- WebAuthn's per-origin credential model
  doesn't by itself imply anything about how a second account would be
  provisioned or isolated.

## Citations

- `openspec/changes/archive/2026-09-16-add-pipeline-dashboard/design.md`,
  "Auth: WebAuthn only, no password path."
- `openspec/changes/archive/2026-09-18-support-multiple-passkeys/`.

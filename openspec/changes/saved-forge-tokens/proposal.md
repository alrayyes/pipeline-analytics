# Proposal

## Why

The register dialog asks for the forge token again whenever the browser
forgets it, because #72 kept it in `localStorage`, which clears with site data
and is readable by any script on the page. Tracked in #462; this is the
contract, reviewed as the design. The handlers and the front end follow.

## Research

- `localStorage` is readable by page scripts, so a long-lived secret doesn't
  belong there
  ([summary of OWASP guidance](https://frontendchecklist.io/rules/security/token-storage-security),
  [what breaks](https://nhimg.org/faq/what-breaks-when-tokens-are-stored-in-localstorage/)).
- A stored secret in a form is masked and never sent back, and an empty field
  means keep it
  ([masked credential fields](https://help.simpplr.com/en-US/simpplr/article/rUBClIbH-masked-credential-fields-for-integrations-and-app-tile-configurations)).

## Decisions

- **A separate record, not a setting.** `GET /api/settings` returns every
  value in clear, so a token can't live there.
- **Write-only.** Responses carry `tokenMasked` only. Registration and
  discovery with no token use the saved one on the server, so the browser
  never needs the plaintext again.
- **One per forge and instance.** Saving replaces. Deleting a saved token
  doesn't touch repos registered with it: each keeps its own encrypted copy.
- **Session-only**, like API tokens and passkeys. Not an MCP tool.
- **No validation on save.** Registration is what tells you whether it works,
  and discovery and registration already report that failure.

## What Changes

- `GET /api/forge-tokens`, `PUT /api/forge-tokens`,
  `DELETE /api/forge-tokens/{tokenId}`.
- `token` is optional on `POST /api/repos` and `POST /api/repos/discover`;
  omitted means the saved one, or `400 no_saved_token`.
- The handlers and the front end follow. The browser copy goes away there.

## Capabilities

### Modified Capabilities

- `forge-ingestion`: a forge token can be saved and reused.

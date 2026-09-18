# Design

## Context

`add-api-token-auth` (still unmerged, PR #186) hardcodes the TTL in
`internal/auth/sqlite/store.go`:

```go
const apiTokenTTL = 365 * 24 * time.Hour
```

used directly in `Store.CreateToken`'s single `INSERT` (`ExpiresAt:
now.Add(apiTokenTTL)`), which takes no TTL parameter today.
`Service.IssueToken(ctx, userID)` and the HTTP handler `issueToken`
(`internal/httpserver/auth.go`) both pass nothing through — there's no
plumbing anywhere for a caller-supplied value. See proposal.md for why
this needs to change; this covers how.

No frontend exists to change: #186 touches no `web/` files, and no
existing route hosts token management.

## Goals / Non-Goals

**Goals:**

- A requested TTL flows from the HTTP request through `Service` to
  `Store`, clamped at the boundary closest to persistence so no caller
  (including a future one) can bypass the ceiling by calling `Service`
  directly.
- Reuse `apiTokenTTL`'s existing role as the ceiling rather than
  introducing a second "max" constant that could drift from it.

**Non-Goals:**

- A token-listing endpoint (unchanged from #178's own non-goal — no
  acceptance criterion here needs one either).
- Nesting this page under `persist-account-settings`'s proposed
  `/settings` route — that change is itself still unimplemented (an
  active, not-yet-built OpenSpec change), and coupling this one to it
  would block token creation on unrelated settings-page work landing
  first. See "Alternatives considered" below.
- Changing anything about how tokens authenticate requests, or the
  session-only gating on issuance/revocation — both unchanged from
  #178.

## Decisions

- **Rename `apiTokenTTL` to `apiTokenTTLCeiling` and add
  `apiTokenTTLDefault = 90 * 24 * time.Hour`.** The existing name reads
  as "the TTL," which stops being true once it's a ceiling a request can
  come in under. Ninety days matches the ticket's requested default
  (alrayyes/pipeline-analytics#191) and the picker's default preset —
  one value, not two that have to be kept in sync.
- **`Store.CreateToken` takes a `requestedTTL time.Duration` parameter
  and clamps it itself**, rather than clamping in `Service` and passing
  an already-final `expiresAt` down. Keeping the clamp at the same layer
  that owns `apiTokenTTLCeiling` today means the ceiling has exactly one
  place it's applied, and a zero/unset `requestedTTL` falls back to
  `apiTokenTTLDefault` in the same spot. `Service.IssueToken` becomes a
  pass-through (`userID string, requestedTTL time.Duration`), matching
  its existing role as a thin layer over the store for this operation.
- **Request field is `ttlSeconds *int` (optional), not an enum of preset
  strings.** The server doesn't need to know about "30 days" vs "90
  days" as concepts — it only needs a duration to clamp, and a raw
  number is simpler to validate (`<= 0` is rejected, everything else
  clamps) than a closed set the server has to keep in sync with
  whatever presets the frontend ships. The frontend's presets are a UI
  choice (proposal.md), not a server-side contract.
- **The picker is its own top-level route (`/tokens`), not nested under
  a future settings page.** See "Alternatives considered."
- **Live expiry preview is computed client-side** (`now + selected
  preset`), not fetched from the server per selection — the ceiling and
  default are static per this design, so there's nothing server-side to
  round-trip for a preview; the server clamp is still authoritative for
  the actual creation request.

## Alternatives considered

- **Nest the token page under `persist-account-settings`'s planned
  `/settings` route.** Rejected: that change is proposed but not built
  (no `/settings` route exists in code yet), and #191 doesn't depend on
  it. Building this as its own route now, and folding it under
  `/settings` later as a small follow-up move if/when that route exists,
  avoids an artificial cross-change dependency neither ticket asked for.
- **Enum of preset strings (`"30d" | "90d" | "1y"`) as the request
  field**, keeping the server preset-aware. Rejected in favor of a raw
  duration: it pushes the "what presets exist" decision into the API
  contract, so adding a fifth preset later would be a backend change
  instead of a frontend-only one.

## Risks / Trade-offs

- Renaming `apiTokenTTL` touches every reference to it
  (`store.go`'s `CreateToken`, its own doc comment, and the two tests in
  `store_test.go` that assert against it) — mechanical, but every call
  site needs updating in the same commit as the rename, not split
  across it.
- A raw `ttlSeconds` field means a client could request `1` second. Not
  a defect: the clamp only enforces a ceiling, not a floor, and a
  user issuing a 1-second token is only shooting their own foot — the
  UI never offers anything close to that, and there's no security
  boundary being protected on the low end the way there is on the
  ceiling.

## Migration Plan

No data migration — `api_tokens.expires_at` already stores an absolute
timestamp; only what computes it before the `INSERT` changes. Existing
rows (none yet, since #186 hasn't merged) are unaffected either way.

## Sequencing

This change's own spec delta (`specs/dashboard-auth/spec.md`) is written
as `ADDED Requirements`, duplicating `add-api-token-auth`'s "API token
issuance and revocation" requirement text rather than `MODIFIED`,
because that requirement doesn't exist in `openspec/specs/dashboard-auth`
yet — only in #186's own pending delta — and OpenSpec's `MODIFIED`
workflow requires the target requirement to already be present in the
archived main spec.

Practically, this means:

- This change should implement **after** #186 merges (its code is what
  the TTL plumbing above modifies) but its planning artifacts don't need
  to wait — nothing here is blocked on #186's exact final shape, since
  the touched functions/signatures are already stable on that branch.
- Once #186 archives (moving its delta into the main spec), this
  change's own delta needs a one-time edit: drop the issuance/
  revocation scenarios duplicated from #186 and re-file the remaining
  TTL-specific scenarios as `MODIFIED` against the now-real main-spec
  requirement. Flagging this here so it isn't missed at archive time for
  *this* change — do this edit before running `openspec archive` on
  `add-api-token-ttl-picker`.

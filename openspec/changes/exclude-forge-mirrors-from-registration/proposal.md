# Proposal

## Why

`ListAccessibleRepos` already excludes archived, forked, and mirror
repos from Discover's results (PR #125), but that check only runs in
the picker's listing path. Nothing stops a mirror, fork, or archived
repo from being registered directly — including one Discover itself
still surfaced, if the forge's bulk repo-listing endpoint doesn't
populate those flags as reliably as a single-repo lookup would (a
known class of quirk in Gitea-derived APIs, unconfirmed here but
plausible given 42 known Forgejo mirrors of GitHub repos got tracked
despite the existing filter). A mirror has no Actions runs of its own
to reconcile — tracking one produces nothing but permanent
reconciliation failures.

## What Changes

- Registration itself (not just Discover's listing) verifies a repo
  isn't archived, a fork, or a mirror before creating it — checked
  against a fresh, single-repo lookup at registration time, not
  reused from whatever Discover returned. This closes the gap
  regardless of whether it's a stale Discover result, a manually-typed
  identifier, or a repo that changed status after being listed.
- Registering an archived/fork/mirror repo is rejected with a message
  saying why, the same way an already-tracked repo is rejected
  (`fix-repo-registration-duplicate-detection`).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `forge-ingestion`: "Repo tracking registration" gains a requirement
  that registration itself, not only the discovery picker, rejects an
  archived, forked, or mirror repository.

## Impact

- Backend: `ingestion.ForgeClient` gains a single-repo lookup method
  (`GetRepo` or similar) both forge clients (`internal/ingestion/github`,
  `internal/ingestion/forgejo`) implement; `Registrar.Register` calls
  it before `store.CreateRepo` and rejects archived/fork/mirror.
- Not included in this change: automatically untracking the 42
  already-registered mirror repos found live. That's a one-time
  manual cleanup (existing `DELETE /api/repos/{repoId}` already
  supports untracking one at a time), not new product behavior.

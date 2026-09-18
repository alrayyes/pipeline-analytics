# Proposal

## Why

`CHANGELOG.md` and the GitHub release notes for `v0.29.0` and `v0.28.1`
each list the same shipped change twice. GitHub's merge-commit strategy
puts both the original branch commit and a synthetic merge commit
(whose body repeats the PR title) into `main`'s history; release-please
reads both as separate conventional-commit entries. Tracked as
[alrayyes/pipeline-analytics#179](https://github.com/alrayyes/pipeline-analytics/issues/179).

## What Changes

- Restrict the repo's merge button to squash-only (disable "create a
  merge commit" and "rebase and merge"), so every future merge produces
  a single commit and release-please never sees the duplicate again.
- Deduplicate the already-published `CHANGELOG.md` sections for
  `0.29.0` and `0.28.1`.
- Deduplicate the already-published GitHub release notes for `v0.29.0`
  and `v0.28.1` to match.

## Capabilities

None of the four existing capabilities (`dashboard-auth`,
`dashboard-ui`, `forge-ingestion`, `pipeline-metrics`) cover release
tooling — they're all about the dashboard's own product behavior. This
introduces a narrowly-scoped capability for the release process itself.

### New Capabilities

- `release-process`: how merges land and how the changelog/release
  notes get generated from them, so a shipped change appears exactly
  once.

### Modified Capabilities

(none)

## Impact

- Repo settings: `mergeCommitAllowed` and `rebaseMergeAllowed` → `false`
  via the GitHub API/`gh api`.
- `CHANGELOG.md` (edited in place for the two affected sections).
- GitHub releases `v0.29.0` and `v0.28.1` (notes edited via
  `gh release edit`).
- No application code, no database, no user-facing behavior.

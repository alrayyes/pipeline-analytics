# Proposal

## Why

`ci.yml` ran all 15 checks on every push and pull request, including
Playwright, Lighthouse and a Docker build, whatever changed. A README edit
paid for the Go test suite, and the hosted-runner queue is shared across
repos, so release pull requests waited behind runs that couldn't fail.
Tracked in #371; follows the account rule "run a check only when the files it
covers change".

## What Changes

- A first `changes` job turns the files a push or pull request touched into
  one boolean per kind of check, and each later job runs only when its group
  is true.
- The groups and their paths live in `.github/scripts/changes.sh`, with a
  test (`changes.test.sh`) that runs inside the `changes` job.
- Editing the workflow or the filter script turns every group on, and an
  unknown base (a new branch) does the same.
- The API docs deploy runs only when the spec or the docs page changed.
- Not filtered, and unchanged: the release, PR-title lint, release and
  Dependabot auto-merge, and SDK-notify workflows.

## Capabilities

### New Capabilities

- `ci-pipeline`: which checks run for which changes.

### Modified Capabilities

(none)

## Impact

- `.github/workflows/ci.yml` (one new job, an `if:` and `needs:` on the rest),
  `.github/scripts/changes.sh` and `changes.test.sh`.
- Job names are unchanged, so main's required status checks still match.
- `CONTRIBUTING.md` says where to add a path.
- A follow-up aligns `lefthook.yml`'s globs with these groups.

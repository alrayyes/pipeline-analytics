# Proposal

## Why

The release's screenshot job opens `docs: refresh README screenshots` and
stops, so the pull request waits for someone to merge it by hand (#639). It
also opens the PR with `GITHUB_TOKEN`, which starts none of the
`pull_request` workflows, so an armed auto-merge would have no checks to wait
for.

## What Changes

- The job checks out and opens the pull request with `RELEASE_TOKEN`, as the
  release PR's auto-merge already does, so a push to the branch and a new PR
  both start the checks. With the default token a refresh PR left open from an
  earlier release would get a new head with no checks, and the required checks
  would keep it blocked.
- It then arms squash auto-merge on it, whether it opened the PR or found one
  already open.
- The PR text says what it is, rather than asking for a manual review. The
  diff is only `docs/` and `README.md`, the paths the job stages.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: the screenshot refresh pull request merges on green.

## Impact

- `.github/workflows/release.yml`.

## Known limit

The repo requires branches to be up to date with `main` before they merge.
Auto-merge does not update a branch, so a refresh PR that waits while `main`
moves needs its branch updated, which this change does not do.

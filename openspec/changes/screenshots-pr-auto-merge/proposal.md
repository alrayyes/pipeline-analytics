# Proposal

## Why

The release's screenshot job opens `docs: refresh README screenshots` and
stops, so the pull request waits for someone to merge it by hand (#639). It
also opens the PR with `GITHUB_TOKEN`, which starts none of the
`pull_request` workflows, so an armed auto-merge would have no checks to wait
for.

## What Changes

- The job opens the pull request with `RELEASE_TOKEN`, as the release PR's
  auto-merge already does, so its checks run.
- It then arms squash auto-merge on it, whether it opened the PR or found one
  already open.
- The PR text says what it is, rather than asking for a manual review. The
  diff is only `docs/` and `README.md`, the paths the job stages.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: the screenshot refresh pull request merges on green.

## Impact

- `.github/workflows/release.yml`.

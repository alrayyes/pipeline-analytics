# Tasks

## 1. Stop future duplicates

- [x] 1.1 Restrict the repo to squash-only merges and verify with
      `gh repo view --json squashMergeAllowed,mergeCommitAllowed,rebaseMergeAllowed`
      that only `squashMergeAllowed` is `true`

## 2. Correct the published record

- [x] 2.1 Deduplicate the `0.29.0` section of `CHANGELOG.md`, keeping
      the entry that references the issue (`closes #175`), and verify
      by reading the section back
- [x] 2.2 Deduplicate the `0.28.1` section of `CHANGELOG.md` the same
      way (two duplicated lines: `e2e` and `nav`), and verify by
      reading the section back
- [x] 2.3 Update the `v0.29.0` GitHub release notes to match via
      `gh release edit`, and verify with `gh release view v0.29.0`
- [x] 2.4 Update the `v0.28.1` GitHub release notes to match via
      `gh release edit`, and verify with `gh release view v0.28.1`

## 3. Ship it

- [ ] 3.1 Open a pull request with `Closes #179` and verify CI passes

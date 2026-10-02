# Tasks

## 1. Filter

- [x] 1.1 `changes.test.sh`: a case per path-to-group mapping, mixed
      changes, no files, pipeline edits, and `--all`
- [x] 1.2 `changes.sh` making the test pass; shellcheck clean

## 2. Workflow

- [x] 2.1 `changes` job: checkout with history, run the test, compute from
      `git diff`, fall back to everything on an unknown base
- [x] 2.2 `needs: changes` and an `if:` on every other job; the Pages deploy
      keeps its spec dependency
- [x] 2.3 Verified on a docs-only pull request (opened on this branch,
      closed unmerged): `changes` and the prose and markdown jobs ran, all 11
      code jobs and the Pages deploy were skipped, status not empty

## 3. Wrap up

- [x] 3.1 `CONTRIBUTING.md` says where to add a path
- [ ] 3.2 Follow-up issue: align `lefthook.yml`'s globs with these groups
- [ ] 3.3 Archive this change once merged

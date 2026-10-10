# Tasks

## 1. Guard

- [x] 1.1 `hook-guard.test.sh`: spec-only, `go.sum`, changelog, docs-only
      and unknown-group cases
- [x] 1.2 `hook-guard.sh` making the test pass
- [x] 1.3 Test runs in the `changes` job

## 2. Hooks

- [x] 2.1 Route the pre-push jobs and the pre-commit `docker build` through
      the guard, dropping their globs
- [x] 2.2 Verified `lefthook`'s guard against a docs-only, a Go-only and a
      spec-only change
- [x] 2.3 `CONTRIBUTING.md` says how hooks pick their files

## 3. Wrap up

- [x] 3.1 Archive this change once merged

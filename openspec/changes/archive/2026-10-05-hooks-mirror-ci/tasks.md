# Tasks

## 1. Parity test

- [x] 1.1 `lefthook-parity.test.sh` fails on each missing hook, the missing
      cache directories and the missing `output: [failure]`
- [x] 1.2 Test runs in the `changes` job

## 2. Hooks

- [x] 2.1 hadolint, `goreleaser check`, Redocly and `go.mod` formatting
      `pre-push` hooks
- [x] 2.2 Cache directories created before each Go hook's Docker run
- [x] 2.3 `output: [failure]` and the CI-only comment
- [x] 2.4 Verified by running each hook against a broken file
- [x] 2.5 `CONTRIBUTING.md` true again

## 3. Wrap up

- [ ] 3.1 Archive this change once merged

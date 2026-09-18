# Tasks

## 1. Forge clients: single-repo lookup

- [x] 1.1 Add `GetRepo(ctx, req) (RepoMetadata, error)` to
      `ForgeClient` (archived/fork/mirror flags at minimum)
- [x] 1.2 Implement it in `internal/ingestion/github`, with a test
      covering an archived/fork/mirror repo and a normal one
- [x] 1.3 Implement it in `internal/ingestion/forgejo`, same coverage

## 2. Registration guard

- [x] 2.1 Add `ErrRepoArchived`/`ErrRepoFork`/`ErrRepoMirror`
      sentinels and check them in `Registrar.Register` via `GetRepo`
      before `store.CreateRepo`, with a unit test per rejection reason
      plus one for a normal repo still succeeding
- [x] 2.2 Wire the three new sentinels into `register`'s handler using
      the same `409`-style response shape
      `fix-repo-registration-duplicate-detection` adds, and verify
      each reason is reported distinctly

## 3. Cleanup (manual, not blocked on 1-2)

- [ ] 3.1 Untrack the 42 already-registered mirror repos on
      git.higherlearning.eu via the existing untrack endpoint, and
      verify the reconciliation log no longer shows connection-refused
      errors for them -- **not done by this PR**: needs direct access
      to the production instance, which this session doesn't have.
      Ryan to do once this ships, per the issue's own Definition of
      Done ("don't consider the incident closed until they're gone").

## 4. Ship it

- [x] 4.1 Open a pull request with `Closes #198` and verify CI passes

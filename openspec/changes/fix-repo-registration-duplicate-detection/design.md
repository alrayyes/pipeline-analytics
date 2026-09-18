# Design

## Context

See `proposal.md` - Why. Relevant existing state:

- `web/src/routes/repos/+page.svelte`'s `trackedIdentifiers` (line
  158) is `$derived` from `repos`, which `loadRepos()` populates from
  a paginated `/api/repos?limit=20&offset=...` call. It only ever
  reflects whatever page is currently loaded.
- `repos` has `UNIQUE (forge, forgejo_instance_url, identifier)`
  (`internal/db/migrations/00001_init.sql:12`) - **found while
  implementing this change: it only fires for Forgejo rows.**
  `forgejo_instance_url` is `NULL` for every GitHub row, and SQL
  treats every `NULL` as distinct from every other `NULL` in a UNIQUE
  constraint, so two GitHub rows with an identical `(forge,
  identifier)` and both-`NULL` instance URLs don't collide at all -
  confirmed live (`INSERT`ing the same values twice raises no error).
  The 42-Forgejo-repo case this proposal cites is real because
  Forgejo rows do have a non-null instance URL; the identical bug on
  GitHub would silently create a second row instead of erroring.
  Migration `00005_repos_unique_identifier_index.sql` adds a
  `COALESCE(forgejo_instance_url, '')`-based unique index alongside
  the existing (harmless, now largely redundant for Forgejo, dead for
  GitHub) inline constraint, closing the gap without touching the
  column's own nullability.
- `internal/httpserver/repos.go`'s `register` handler
  (lines 84-111) maps every error from `Registrar.Register` to a flat
  `500 internal_error "register repo"`, with zero `slog` calls
  anywhere in the file.
- `internal/ingestion/registrar.go`'s `Register` calls
  `r.store.CreateRepo`, whose only error path today is
  `fmt.Errorf("insert repo: %w", err)` wrapping whatever the raw SQL
  driver error is - there's no typed/sentinel error to distinguish "a
  UNIQUE violation" from any other insert failure.

## Goals / Non-Goals

**Goals:**

- Discover never re-offers an already-tracked repo, regardless of how
  many repos are tracked.
- A duplicate-registration attempt gets a response that says so, and
  a log line, instead of an opaque 500.

**Non-Goals:**

- General pagination redesign for the Repos page. `PAGE_SIZE = 20`
  stays for the page's own rendering; only the "is this already
  tracked" check needs the full set.
- The mirror/fork/archived exclusion itself - that's
  `exclude-forge-mirrors-from-registration`, a separate change. This
  one is purely about duplicate detection and error handling.

## Decisions

**A dedicated identifiers-only query, not "just raise the limit".**
Bumping `PAGE_SIZE` or fetching with a very high limit still has a
ceiling and re-fetches every column (including the masked token) just
to build a set of strings. A new store method
(`ListRepoIdentifiers(ctx, forge, instanceURL) ([]string, error)`, no
pagination) backing a lightweight new query param or endpoint is
worth the small addition - `trackedIdentifiers` only ever needs the
identifier strings, not full `Repo` objects.

**A sentinel error for the constraint violation, detected via SQLite's
error code, not string-matching the driver's message.** **Correction
from this proposal's original wording:** this codebase's actual driver
is `modernc.org/sqlite` (`go.mod`), not `mattn/go-sqlite3` - there's no
CGo binding here to have a `sqlite3.ErrConstraintUnique` from.
`CreateRepo` instead does `errors.As` for `*sqlite.Error` (the driver's
own type, `modernc.org/sqlite`) and checks `Code() ==
sqlite3lib.SQLITE_CONSTRAINT_UNIQUE` (the extended result code, from
`modernc.org/sqlite/lib`), returning a new
`ingestion.ErrRepoAlreadyTracked` sentinel instead of the raw wrapped
error when it's specifically this constraint. The HTTP handler then
does one `errors.Is` check: that sentinel maps to `409`, everything
else stays a logged `500`.

**Every branch of `register` gets a `slog.ErrorContext` call**, not
just the new conflict case - the current total absence of logging
here is itself part of the bug (proposal.md's Why), and leaving the
non-conflict 500 path silent would just relocate the same problem.

## Risks / Trade-offs

- **A second query for identifiers is one more round trip on every
  Discover run** -> negligible: it's a single indexed lookup against
  a table sized to a solo developer's tracked repos, not a
  performance-sensitive path.
- **Detecting the SQLite error code ties `CreateRepo` to a specific
  driver's error type** -> acceptable: this codebase already commits
  to SQLite specifically (`internal/db`, `internal/auth/sqlite`, etc.
  are not written against a driver-agnostic abstraction), so this
  isn't a new coupling, just the same one applied to error handling.

## Migration Plan

1. Migration `00005_repos_unique_identifier_index.sql`: a
   `COALESCE`-based unique index so a GitHub duplicate actually
   violates a constraint (see Context - this codebase's existing
   inline `UNIQUE` never did for GitHub rows).
2. Add `ErrRepoAlreadyTracked` and make `CreateRepo` detect the
   constraint violation (from either index) and return it.
3. Add `ListRepoIdentifiers` (store + a lightweight API surface) and
   switch `trackedIdentifiers` to use it instead of the paginated
   `repos` fetch.
4. Update `register`'s handler: `errors.Is(err,
   ingestion.ErrRepoAlreadyTracked)` -> `409`; everything else stays
   `500`; both paths get a `slog.ErrorContext` call.

Mostly no rollback concern: additive error handling and a read-only
query, no behavior change for anything already working. The one real
migration risk is step 1 on a database that already has duplicate
GitHub rows from this exact bug - `CREATE UNIQUE INDEX` fails outright
against existing duplicates rather than silently ignoring them, which
surfaces the problem instead of masking it, but means step 1 isn't
guaranteed to apply cleanly on every existing deployment without
first de-duplicating any rows it finds.

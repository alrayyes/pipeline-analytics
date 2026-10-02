# Design

## Decisions

**Source: `CHANGELOG.md`, not the GitHub API.** release-please writes a
release's section into the changelog in the release PR, which merges before
the tag. The release workflow then builds the web app from the tagged commit,
so the notes for the release being cut are already in the file the build
reads. That also means no token, no network at build time, and a build that
reproduces from the repository alone.

**A JSON file, loaded by the page, not a bundled import.** `releases.json`
lands in the built frontend (`internal/webassets/dist`) beside the other static
assets, so the Go binary serves it from its embedded copy. The page fetches it same-origin. A missing file
falls through the SPA fallback to `index.html`, which fails to parse and
reaches the existing error state, so a build that skipped the step shows the
message and the GitHub link instead of a blank page.

**Parser in `src/lib`, script in `scripts/`.** The parsing is pure and lives
in `src/lib/changelog.ts` so `bun test` covers it like the other `lib`
modules; `scripts/generate-releases.ts` only reads the file, calls it and
writes the JSON. Bun runs the script directly, matching the rest of the
toolchain.

**Generated, not committed.** `releases.json` comes from `CHANGELOG.md`;
committing it would mean the release PR changes two files that must agree.
It's gitignored and produced by `prebuild`/`predev`.

**Fields.** `version`, `tag` (`v` + version), `name` (same as the tag, as
GitHub names these releases), `url` (the release's GitHub page, derivable from
the tag), `date` (`YYYY-MM-DD`, the only precision the changelog has and the
only one the page shows) and `body` (the section's markdown without its
heading, which the card header already shows).

**The page still renders the markdown at runtime** with `marked` and
DOMPurify, as it did for the API's body, so the rendering and its XSS
handling don't change.

## Risks / Trade-offs

- **A heading format change breaks the parse.** The 79 existing headings share
  one shape. A parser test pins it, and a changelog with no recognisable
  heading yields an empty list instead of throwing, so a format change shows
  "No releases yet" rather than breaking the build. Worth a look if
  release-please's format changes.
- **The page can lag a release by one build.** It shows what was in the
  changelog when this binary was built, which is exactly the release it is.
  Running an older image shows older history, which is correct for it.
- **`published_at` loses its time of day.** The page only ever showed the date.

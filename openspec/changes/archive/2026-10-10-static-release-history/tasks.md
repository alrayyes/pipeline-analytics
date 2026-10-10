# Tasks

## 1. Parser

- [x] 1.1 Failing tests for `parseChangelog`: heading, date, body without the
      heading, ordering, several releases, no headings, a body with its own
      `###` sections
- [x] 1.2 `src/lib/changelog.ts`

## 2. Build step

- [x] 2.1 `scripts/generate-releases.ts` writing `static/releases.json`;
      `prebuild` and `predev` scripts; gitignore the output
- [x] 2.2 Verify `bun run build` emits `releases.json` in the built frontend with every
      release in `CHANGELOG.md`

## 3. Page

- [x] 3.1 Update the Playwright journey first: fail on any `api.github.com`
      request, assert releases render from the bundled file
- [x] 3.2 `/releases` loads `/releases.json`; keep the error state

## 4. Wrap up

- [x] 4.1 Docs true again: the README doesn't mention where release notes
      come from; CONTRIBUTING documents the new build step
- [x] 4.2 Archive this change once merged

# Tasks

## 1. Add the file

- [x] 1.1 Create `web/static/humans.txt` with `/* TEAM */` (GitHub handle
      `alrayyes` only — no email, no other personal details) and
      `/* SITE */` (Go, SvelteKit, SQLite) sections, satisfying
      `specs/site-metadata` — verify it lists no PII beyond the handle
- [x] 1.2 Run `bun run dev` in `web/` and verify `GET /humans.txt` returns
      `200` with a `text/plain` body served as-is (spec scenario:
      "Visitor requests humans.txt")

## 2. Ship it

- [x] 2.1 Open a pull request with `Closes #175` and verify CI passes

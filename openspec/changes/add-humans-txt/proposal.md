# Proposal

## Why

The site has no `humans.txt`, so a visitor curious who built it or what
it's built with has nothing to check short of digging through the repo.
`humans.txt` is a plain-text convention (humanstxt.org) for exactly that.
Tracked as
[alrayyes/pipeline-analytics#175](https://github.com/alrayyes/pipeline-analytics/issues/175).

## What Changes

- Add `web/static/humans.txt`, served the same way `web/static/robots.txt`
  already is — a static file, no new routing code.
- Content follows the humanstxt.org `/* TEAM */` + `/* SITE */` section
  convention: a `TEAM` section naming the maintainer by GitHub handle
  (`alrayyes`), no other personal details; a `SITE` section naming the
  stack (Go, SvelteKit, SQLite).

## Capabilities

None of the existing capabilities (`dashboard-auth`, `dashboard-ui`,
`forge-ingestion`, `pipeline-metrics`) cover static discovery files served
from `web/static/` — they're about presenting ingested pipeline data, not
site metadata. This introduces a new, narrowly-scoped capability instead.

### New Capabilities

- `site-metadata`: static discovery files the web app serves about
  itself (starting with `humans.txt`), separate from the dashboard's own
  pipeline-data capabilities.

### Modified Capabilities

(none)

## Impact

- `web/static/humans.txt` (new file).
- No code, routing, or dependency changes — SvelteKit already serves
  everything under `web/static/` at the site root.

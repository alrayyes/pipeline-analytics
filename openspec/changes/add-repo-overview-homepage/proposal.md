# Proposal

## Why

The login landing page is a flat, cross-repo list of every pipeline,
grouped by repo as inline sections on one page. Once more than a
handful of repos are tracked, scanning that list for "what's actually
broken right now" means reading past every healthy pipeline first. A
repo-first landing page — each repo shown once with its own health
rollup, sorted so the unhealthy ones surface first — turns the same
data into an actual triage view, the same shape as GitHub's own
Actions overview or a service-health dashboard.

## What Changes

- `/` becomes a repo overview: each tracked repo is a card showing an
  aggregate health status and the names of its unhealthy pipelines
  (not just a count), with the same all/healthy/unhealthy filter the
  current homepage has, defaulting to unhealthy.
- A new per-repo pipeline list at `/repos/[id]` — the same
  card/health-filter/sort UI the current homepage has, scoped to one
  repo (no repo selector, no forge filter — both redundant once
  scoped). Sibling to the existing `/repos/[id]/usage` page.
- **BREAKING**: the flat, cross-repo "all pipelines on one page" view
  (today's `/`) and its "Pipelines" nav link are removed. The repo
  cards (listing unhealthy pipelines by name) plus per-repo drill-down
  cover the same information without a redundant third view.
- **BREAKING**: repo registration/management (today's `/repos` —
  discover, add, untrack) moves into a "Repositories" section on the
  `/settings` page (`persist-account-settings` introduces that page
  for theme/filter preferences; this change adds the second section to
  it). The `/repos` route and its "Repositories" nav link are removed
  in favor of a "Settings" nav link.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: "Repo/pipeline overview" changes from a flat
  cross-repo pipeline list to a repo-first landing page; adds a new
  requirement for the per-repo pipeline list at `/repos/[id]`.

## Impact

- Frontend: `web/src/routes/+page.svelte` rewritten from a flat
  pipeline list to a repo-card overview; new
  `web/src/routes/repos/[id]/+page.svelte`; `web/src/routes/repos/
+page.svelte`'s discover/add/untrack UI moves into the `/settings`
  page's new "Repositories" section; `web/src/lib/components/Nav.svelte`
  loses the "Pipelines" and "Repositories" links, gains "Settings".
- Depends on `persist-account-settings` existing (the `/settings`
  route and page shell) — this change adds a section to that page
  rather than creating it.
- API: no new endpoints — `/api/repos` and `/api/pipelines` already
  return what both pages need; this is a client-side reshaping of
  existing data.

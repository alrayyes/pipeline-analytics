# Design

## Context

See `proposal.md` - Why. Relevant existing state:

- `web/src/routes/+page.svelte` already fetches the full,
  unpaginated `/api/pipelines` and `/api/repos` and groups pipelines
  by repo client-side (`repoById`, `groupedPipelines`) - there is no
  server-side grouping or per-repo endpoint today, and the existing
  code comments explicitly call this out as deliberate ("already does
  a full unpaginated fetch ... so there's no server round trip to
  save"). Both new pages keep that pattern rather than adding a
  `?repoId=` query param to `/api/pipelines`.
- `web/src/routes/repos/[id]/usage/+page.svelte` establishes the
  `/repos/[id]/...` nested-route shape already; the new per-repo
  pipeline list is a sibling `+page.svelte` at that same `[id]`
  segment's index, not a new pattern.
- `web/src/routes/repos/+page.svelte` (discover/add/untrack) is 762
  lines - moving it into `/settings` means relocating it as its own
  component, not inlining it into the settings page file.
- This change assumes `persist-account-settings`'s `/settings` page
  shell already exists to add a section to. Neither change has been
  implemented yet; if this one lands first, its "Repositories" section
  becomes the first thing on that page instead of the second.

## Goals / Non-Goals

**Goals:**

- Repo overview surfaces every unhealthy pipeline by name, at a
  glance, from the landing page - no regression against
  `dashboard-ui`'s existing "no per-item click-through" requirement.
- No new API endpoints; reshape existing data client-side, consistent
  with how filtering already works elsewhere in this app.

**Non-Goals:**

- Changing `/api/pipelines` or `/api/repos`'s response shape.
- Changing the per-pipeline detail view (`/pipelines/[id]`), the Steps
  page, or the Usage page - none of them are touched by this change.
- Pagination for the repo overview. The existing pipelines page has
  none either, and this app's scale (a solo developer's tracked repos)
  doesn't need it yet.

## Decisions

**Repo health rollup: worst-of aggregation, computed client-side.**
A repo's card status is `unhealthy` if any of its pipelines are
`unhealthy`, else `healthy`, else (no pipelines ingested yet) a
neutral "no runs yet" state - reusing the existing empty-state message
pattern (`pipelines.length === 0` branch already in
`+page.svelte`). No new status value; the domain only ever has two.

**Repo card lists unhealthy pipelines by name, not just a count.**
Directly required by `dashboard-ui`'s unchanged "without requiring the
user to open each pipeline individually" wording - a count alone would
regress it. When the health filter is `all`, a healthy repo's card
shows a compact "N healthy" line instead of naming every pipeline,
since there's nothing actionable to name.

**Extract the pipeline-card list (health filter + sort + cards) into
a shared component**, used by the new `/repos/[id]` page essentially
unchanged from today's homepage body, minus the repo selector and
`ForgeFilter` (both redundant once a single repo is already the
route's scope). The repo overview page does _not_ use this component -
its card is a different, denser summary view, not a full pipeline
list.

**Repo management UI moves as a component, not inlined.** Extract
today's `web/src/routes/repos/+page.svelte` body into
`RepoManagement.svelte` and mount it as a section on `/settings`,
rather than growing the settings page file to 762+ lines directly.

## Risks / Trade-offs

- **Removing the flat cross-repo Pipelines page is a real feature
  removal** (marked BREAKING in the proposal) -> anyone with that URL
  bookmarked gets a 404 once `/` changes shape. Accepted: this is a
  single-user personal tool, and the repo overview plus per-repo
  drill-down cover the same information one click deeper.
- **Sequencing with `persist-account-settings`** -> if this change's
  implementation starts before that one's `/settings` shell exists,
  the "Repositories" section has nowhere to land. Mitigated by calling
  this out as a real dependency on the tracking issue (forge
  dependency link), not just a mention in prose.

## Migration Plan

1. Add the repo overview page at `/` (replacing the flat pipeline
   list) and the repo health-rollup logic.
2. Add `web/src/routes/repos/[id]/+page.svelte` (repo-scoped pipeline
   list), extracting the shared pipeline-card-list component from
   today's homepage body in the process.
3. Extract `RepoManagement.svelte` from today's `/repos` page and
   mount it as a section on `/settings` (once that route exists).
4. Remove `web/src/routes/repos/+page.svelte` and the old flat
   `/` pipeline-list body; update `Nav.svelte`'s links (drop
   "Pipelines"/"Repositories", add "Settings").

No rollback concern beyond reverting the commit: nothing server-side
changes, so there's no data migration to undo.

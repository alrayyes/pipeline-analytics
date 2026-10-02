# Design

## Context

Existing state this builds on (verified against the tree):

- `runs`, `jobs`, `steps` tables carry status, conclusion and
  timestamps (`00001_init.sql`). `runs` has no branch, SHA, commit
  message or actor.
- `GET /api/runs/{runId}/steps` returns one run's steps;
  `GET /api/pipelines/{id}/flaky-runs` returns failed runs for one
  named step. There is no run *list* and no cross-pipeline aggregate.
- `GET /api/steps/unhealthy` already groups flaky/failing steps by
  pipeline (`/steps` page).
- Web: SvelteKit static adapter, shadcn-svelte + Tailwind 4, layerchart
  for charts, `$lib/dashboardApi.ts` as the shared fetch layer,
  `*.svelte.ts` rune stores for filters (`forgeFilter`,
  `pipelinesFilters`).

## Mapping: screen to route, data, components

| Design screen | Route | Data | Notes |
| --- | --- | --- | --- |
| CI/CD Failure Overview | `/` (replaces the repo overview) | `GET /api/insights/failures` | Window toggle, repo scope |
| Pipelines & Runs | `/runs` | `GET /api/runs` | Existing `/runs/[id]` stays as the run detail |
| Root Cause Diagnostics | `/failures` | `GET /api/insights/failures` (`failureGroups`) | Links into `/pipelines/[id]/flaky-runs` |
| Flaky Tests & Telemetry | `/flaky` | `GET /api/steps/unhealthy` plus per-step history | Supersedes `/steps`, which redirects here |

## Components (`web/src/lib/components/telemetry/`)

- `MetricCard`: label, value, delta, status tone. Used by overview
  and flaky header.
- `StageProgress`: segmented bar of a run's steps; each segment is
  pass/fail/running/skipped/pending. Pure presentational, takes
  `steps: RunStep[]`.
- `FlakeMatrix`: grid of the last N runs of a step, pass/flake cells,
  with an accessible text summary ("41 flakes in 120 runs").
- `StageDistribution`: donut or bar of failures by step name
  (layerchart), legend as a list so it is readable without colour.
- `FailureGroupCard`: root-cause entry: step name, category chip,
  occurrence count, affected pipelines, "view failed runs" link.
- `RunCard`: status, pipeline, `StageProgress`, duration, commit tag
  (mono SHA), actor, forge deep link.
- `StatusBadge`: one tone map: pass green, fail red, flaky amber,
  running cyan. Colour is never the only signal (icon + text).
- `WindowToggle`: 24h/7d/30d, backed by the `Window` API parameter.
- `TabBar`: bottom navigation at phone width; `Nav.svelte` keeps the
  desktop layout.

## State

Following the existing `*.svelte.ts` pattern, persisted through
`/api/settings` where it already is:

- `telemetryWindow` (`24h|7d|30d`, default `7d`), shared by overview,
  failures and flaky views. Persisted like `pipelinesFilters`.
- `runsFilters`: `status` (`all|failed|running|success`), `repoId`,
  `limit/offset`. Resets to the first page on change (matches the
  existing pagination scenario).
- Fetched data stays in page-local `$state`, as today; no global
  cache.
- Run list polls every 15s while any listed run is `running`, stops
  otherwise. Replaces the design's "Live Sync" label with real
  behaviour.

## Decisions

**Stage means step.** Forges expose workflows, jobs and steps; there
is no stage taxonomy. "Failure stage distribution" groups by failing
step name; the progression bar renders a run's steps.

**Failure categories are heuristic and labelled as such.** The root
cause view's buckets (Infra/OOM, Code/Tests, Net/Timeouts, Config/
Secrets) come from step conclusion (`timed_out`, `cancelled`,
`failure`) and a small name/exit-code matcher, with an "Uncategorised"
bucket. No log parsing. The matcher lives in Go beside the metrics
code so the API owns it and it is table-tested.

**MTTR is derived, not stored**: time from a pipeline's first failed
run to its next successful run, averaged over recoveries in the
window; absent when there were none.

**No write actions.** Buttons in the designs that mutate a forge are
dropped; "Create Issue"-style actions are not offered. A deep link to
the forge replaces them.

**Check for an SDK**: forge calls stay behind the existing ingestion
clients; no new integration is added.

**Spec first**: `openapi/openapi.yaml` gains both endpoints and is
linted with Redocly before handlers exist.

**Tests**: Playwright journey per view with an axe-core scan in the
same test; unit tests for the category matcher, MTTR and stage
mapping; component tests for `StageProgress` and `FlakeMatrix`
states.

**Theme**: Telemetry Dark's `#0f141b` background and status colours
map onto existing tokens (`--background`, `--destructive`, chart
tokens), plus `--warning` and `--running` added to `app.css`. Light
mode keeps working; the app's theme toggle is unchanged. Biome
`useBaseline` and the Tailwind lint apply as usual.

## Risks / Trade-offs

- **Heuristic categories can mislabel.** Mitigated by the visible
  "Uncategorised" bucket and by showing the raw step conclusion on
  each group.
- **Aggregates over a full window scan.** Single-user scale; the
  `idx_runs_repo_pipeline` index covers the filters. Revisit if a
  query exceeds 200ms on the largest tracked repo.
- **Commit fields backfill.** Old runs show no commit tag. No
  backfill; reconciliation fills runs it re-fetches.
- **Scope creep from the designs.** The design shows more than the
  data supports; every dropped element is listed in the proposal so
  the omission is deliberate.

## Settled questions

1. **The failure overview replaces `/`.** The repo overview cards
   move to `/repos`; the overview's repo scope covers the same triage
   need. This is **BREAKING** for anyone bookmarking the old landing
   page's shape.
2. **`/flaky` replaces `/steps`**, with a redirect, since both list
   flaky and failing steps.
3. **The bottom tab bar shows at phone widths only.** Desktop keeps
   `Nav.svelte`'s layout.

# Proposal

## Why

The dashboard shows per-pipeline health and per-step breakdowns, but
not the question a broken build actually raises on a phone: what is
failing across everything, how bad is it, and where in the run did it
break. Four mobile designs in Stitch ("Telemetry Dark") answer that:
a failure overview, a run list with stage progression, a root-cause
view grouped by failing step, and a flaky-test view. Tracked in #335.

Source: Stitch project `projects/9077764458780127669`, screens
`CI/CD Failure Overview` (`56944b30811741c8b3242a892111fc20`),
`Pipelines & Runs` (`d0bdef175b8145e4ad1ccc8e5994bb0e`),
`Root Cause Diagnostics` (`9a5e9d2c194b4071a4da5d52df3a77a8`) and
`Flaky Tests & Telemetry` (`6ea2bad1c67a4044b97c6833c3ec8895`).

## What Changes

Four views, each limited to what ingested data can feed:

- **Failure overview**: pass rate (with change vs. the previous
  window), failed-run count, flaky-step ratio, MTTR, a failure-stage
  distribution, and the top failing pipelines, for a 24h/7d/30d
  window and an optional repo scope.
- **Pipelines & runs**: a filterable run list (all/failed/running/
  success). Each run shows a stage progression bar built from its
  recorded steps, status, duration, commit and actor, and a forge
  deep link.
- **Root cause diagnostics**: failures grouped by failing step with
  occurrence counts, affected pipelines and failure-category buckets
  inferred from step conclusion and name, each linking to its failed
  runs.
- **Flaky tests & telemetry**: per flaky step, a pass/fail matrix of
  its last 40 runs, flake rate and run count, ranked by flake rate.
- New read endpoints: `GET /api/insights/failures`, `GET /api/runs`.
- Ingestion starts recording run branch, head SHA, commit message and
  actor so runs can carry the commit tags the designs show.

### Out of scope (follow-up issues)

- Anything that writes to a forge: Re-run, Cancel, Quarantine, Apply
  PR, File Jira, Auto-Remediation Rules. `dashboard-ui` stays
  read-only toward forges.
- Raw log retrieval and the live ANSI terminal. No logs are ingested;
  designs' log snippets are replaced by a deep link to the step's log
  on the forge.
- AI fix prescriptions.
- Runner/cluster data (autoscaler, active runners, Kubernetes pods).
- Per-branch environment selector ("production / main"): runs gain a
  branch field here, but selecting by branch is a follow-up.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds the four views and their navigation.
- `pipeline-metrics`: adds window aggregates (pass rate, MTTR,
  failure-stage distribution, flake ratio, failure-category
  grouping).
- `forge-ingestion`: records run branch, SHA, commit message, actor.

## Impact

- Frontend: new routes, shared telemetry components (see design),
  `Nav.svelte` becomes a bottom tab bar on phone widths.
- Backend: migration `00008` adds four nullable columns to `runs`;
  both webhook and reconciliation paths populate them; two new
  handlers plus OpenAPI entries (spec first, linted with Redocly).
- Existing rows keep null commit fields; the UI omits the tags.
- Theme: the "Telemetry Dark" palette maps onto the existing
  shadcn tokens in `web/src/app.css` rather than adding a second
  theme.

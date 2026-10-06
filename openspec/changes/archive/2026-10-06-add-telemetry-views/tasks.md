# Tasks

## 1. API contract (spec first)

- [x] 1.1 Add `GET /api/insights/failures` and `GET /api/runs` to
      `openapi/openapi.yaml`; lint with Redocly
- [x] 1.2 Review the contract before any handler is written

## 2. Ingestion

- [x] 2.1 Failing test: run webhook stores branch, SHA, message,
      actor; migration `00008` adds the nullable columns
- [x] 2.2 Populate the same fields in reconciliation, GitHub and
      Forgejo, verified against both forges' payload shapes

## 3. Metrics

- [x] 3.1 Window aggregates (pass rate, delta, flake ratio) with
      table tests
- [x] 3.2 MTTR with tests for no-recovery and multiple recoveries
- [x] 3.3 Step/category grouping and the category matcher, table
      tested including "uncategorised"
- [x] 3.4 Handlers for both endpoints

## 4. Frontend foundations

- [x] 4.1 `dashboardApi.ts` types and fetchers, the status model, stage
      mapping and the `telemetryWindow` store, all with unit tests
      (`runsFilters` is page-local and lands with the runs view, 5.2)
- [x] 4.2 `running` and `flaky` badge variants and `StatusBadge`, using the
      existing Tailwind palette like the `success` variant instead of new
      theme tokens; contrast is checked by the axe scan in light and dark
- [x] 4.3 Shared components, each landing with the first view that uses it
      and covered by that view's Playwright journey, since this repo has no
      component-level test layer: `StageProgress` and `RunCard` (done, 5.2);
      `TabBar` (done, 5.1c); `FlakeMatrix` (done, 5.4c); `MetricCard` (done, 5.1b; the stage
      distribution is inline on the overview, not a component);
      `WindowToggle`, `CategoryBreakdown` and `FailureGroupCard` (done, 5.3)

## 5. Views (one pull request each)

- [x] 5.1a Move the Pipelines list to `/pipelines` and redirect `/` there.
      (The "repo cards" planned for `/repos` never existed: `add-repo-overview-
      homepage` was not implemented, so `/` was the flat Pipelines list.)
- [x] 5.1b Failure overview at `/`
- [x] 5.1c Phone-width bottom tab bar
- [x] 5.2 Runs list with polling
- [x] 5.3 Root cause diagnostics
- [x] 5.4 Flaky telemetry; `/steps` redirects to `/flaky`
  - [x] 5.4a `GET /api/steps/flaky` in `openapi/openapi.yaml`: flaky steps
        across pipelines ranked by flake rate, each with its run count and
        up to 40 recent outcomes (the unhealthy list has neither)
  - [x] 5.4b Metrics, handler and the `list_flaky_steps` MCP tool, tests
        first; drop the `notMCPTools` entry for it
  - [x] 5.4c The `/flaky` view and `FlakeMatrix`
  - [x] 5.4d `/steps` redirects to `/flaky`; the old Steps page and its
        journey go
- [x] 5.5 Playwright journey and axe scan per view at 390px (the
      `overview`, `runs`, `failures` and `flaky` specs each carry theirs,
      light and dark)

## 6. Docs

- [x] 6.1 README and screenshots (`docs/`) true again; `ARCHITECTURE.md`
      checked and unchanged, it names no view
- [x] 6.2 Archive this change

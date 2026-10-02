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
- [ ] 3.4 Handlers for both endpoints

## 4. Frontend foundations

- [ ] 4.1 `dashboardApi.ts` types and fetchers (types, fetchers, status model and stage mapping done); `telemetryWindow` and
      `runsFilters` stores with unit tests
- [ ] 4.2 Theme tokens (`--warning`, `--running`) and `StatusBadge`
- [ ] 4.3 `MetricCard`, `StageProgress`, `FlakeMatrix`,
      `StageDistribution`, `FailureGroupCard`, `RunCard`,
      `WindowToggle`, `TabBar` with component tests

## 5. Views (one pull request each)

- [ ] 5.1 Failure overview at `/`; move the repo cards to `/repos`
- [ ] 5.2 Runs list with polling
- [ ] 5.3 Root cause diagnostics
- [ ] 5.4 Flaky telemetry; `/steps` redirects to `/flaky`
- [ ] 5.5 Playwright journey and axe scan per view at 390px

## 6. Docs

- [ ] 6.1 README, screenshots (`docs/`) and `ARCHITECTURE.md` true
      again; archive this change

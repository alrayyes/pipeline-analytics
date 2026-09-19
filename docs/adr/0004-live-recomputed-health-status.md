# 4. Pipeline health is recomputed live, nothing persisted

## Status

Accepted

## Context

Each pipeline's health status (healthy/unhealthy, with which signals
triggered it) is a derived view over its recent run/step history, not a
fact recorded when a run completes. The obvious alternative -- compute
it once when a run finishes and persist the result -- would look like an
optimization: no recomputation cost on every `GET /api/pipelines`
request.

## Decision

Health status has no stored representation at all. `metrics.Service`
(`internal/metrics/metrics.go`) recomputes it from scratch on every
request, from a sliding window of the most recent runs/steps
(`Window.RunCount`, default 20). There's no cache, no invalidation logic,
and no background job that keeps a persisted status current.

## Consequences

- A pipeline's health status is always consistent with its current
  window of history -- there's no stale-cache class of bug, because
  there's no cache.
- It self-heals with no extra code: once a run that failed or a step
  that flaked ages out of the window, the next request already reports
  the pipeline healthy again, with no explicit "clear the unhealthy
  flag" step anywhere. This is load-bearing for the flaky-step navigation
  feature (#216)
  (`internal/httpserver/pipelines_test.go`'s `TestPipelineHealth_SelfHeals`
  pins this down as a regression test, added when that feature's own
  design depended on it staying true).
- The cost is repaying the same computation on every pipeline-list
  request rather than once per ingested run. Accepted because this is a
  single-account tool with, at most, a handful of tracked repos, and a
  bounded window size, not a scale where that repeated computation is
  actually expensive. Revisit if either assumption stops holding.
- Persisting health status later (for a dashboard with real scale, say)
  would need its own invalidation story -- this ADR is the record that
  the "just cache it" option was available and specifically not taken,
  not an oversight to quietly fix.

## Citations

- `internal/metrics/metrics.go`: `Service.summarizeFromRuns`,
  `computeHealth`.
- `internal/httpserver/pipelines_test.go`: `TestPipelineHealth_SelfHeals`.
- PR #25 (`feat(metrics): pipeline health, trend, ranking, and usage
computation`) established this shape; PR #220 (issue #216's flaky-step
  navigation feature) is what surfaced it as a property worth pinning
  down explicitly.

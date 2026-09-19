# 3. Webhook-primary ingestion with a polling fallback, no GraphQL

## Status

Accepted

## Context

Pipeline run/job/step data needs to reach this service in near-real-time
without burning through GitHub's REST rate limit (5,000 requests/hour
per token, shared across every tracked repo on that token) or adding a
second API paradigm alongside the REST client webhook management already
needs.

Alternative considered: GitHub's GraphQL API, batching multiple repos'
run history into fewer requests. Rejected -- it would reduce request
count, but conditional REST (below) already achieves the actual goal
without a second client/paradigm to maintain. Forgejo has no GraphQL API
at all, so this question doesn't apply there regardless.

## Decision

`workflow_run`/`workflow_job` webhooks carry the real-time load at zero
ongoing API cost per event. A reconciliation poll runs on an interval
(`--reconcile-interval`, default one hour) per tracked repo, to backfill a
newly registered repo's history and catch any webhook delivery that was
missed. On GitHub, reconciliation uses conditional requests
(`If-None-Match` against the response `ETag`): a `304 Not Modified`
response doesn't count against the primary rate limit, provided the same
token is used consistently.

## Consequences

- Losing a webhook delivery (a flaky network blip, GitHub/Forgejo-side
  outage) is self-healing within one reconciliation interval, not a
  silent permanent gap.
- The reconciliation poll's own request pattern has to stay
  conditional-request-clean to hold the "zero ongoing cost" property --
  a change that stops sending `If-None-Match` correctly would quietly
  reintroduce real rate-limit pressure at scale (more tracked repos).
- Two ingestion paths (webhook receiver, reconciliation poller) means two
  places a run/job/step can first land, both writing through the same
  `ingestion.RunStore` port -- kept idempotent (`UpsertRun`/`UpsertJob`)
  so either path observing the same run twice is a no-op, not a
  duplicate.

## Citations

- `openspec/changes/archive/2026-09-16-add-pipeline-dashboard/design.md`,
  "Ingestion: webhooks primary, REST reconciliation fallback, no
  GraphQL."

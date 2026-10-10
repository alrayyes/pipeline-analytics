# Design

## Context

Flakiness is derived on every request from step history: `aggregateSteps`
marks a step flaky when it both passed and failed in the window, and
`computeHealth`, `flakyStepRatio` and `ListFlakySteps` all read that. ADR 0004
records that health is recomputed and never stored, so it self-heals. A
quarantine is the first state a person writes that no ingest produces. See
proposal.md for what it does.

## Goals / Non-Goals

**Goals:**

- Keep health recomputed per request: a quarantine is an input to the
  computation, not a stored status.
- One place decides whether a flaky step counts, so the health signal, the
  ratio, the unhealthy-steps overview and the MCP tools can't disagree.

**Non-Goals:**

- Quarantining a failing, non-flaky step. Only a step that is flaky in the
  window can be marked; a mark on a step that stops flaking is kept until it
  expires and does nothing.
- Per-branch quarantine. A step is one name within one pipeline, as it is for
  `FlakyStep`.
- Any forge write, any notification, any audit trail beyond the mark's own
  timestamps and note.

## Decisions

- **A table, joined at read time.** `step_quarantines(repo_id, pipeline,
  step, note, quarantined_at, expires_at)`, primary key `(repo_id, pipeline,
  step)`. Alternatives: a flag on the step rows (rejected: steps are ingested
  per run, a mark would have to be copied onto every new occurrence) and
  storing a derived health status (rejected by ADR 0004).
- **Expiry is computed, not swept.** A row is active when `expires_at` is after
  now. Nothing deletes expired rows on a timer; an upsert replaces one, and a
  read ignores one. This keeps "no background job" from ADR 0004 true, and
  renewing is the same upsert as creating.
- **The metrics service takes the active set once per request** and passes it
  to `computeHealth`'s `anyFlaky`, `flakyStepRatio` and `ListFlakySteps`
  through one helper, `countsAsFlaky(step, quarantined)`. Alternative: filter
  in SQL (rejected: the flaky definition lives in Go, and a second copy of it
  in a query drifts).
- **`Step.Flaky` stays true for a quarantined step.** The list still ranks and
  shows it; only the callers that feed health and the ratio ask
  `countsAsFlaky`. The new `Quarantined`, `QuarantineNote` and
  `QuarantineExpiresAt` fields ride alongside it.
- **The write route is session-only**, reusing `requireSessionAuth` from
  `forge_tokens.go`: `PUT /api/pipelines/{pipelineId}/steps/{step}/quarantine`
  and `DELETE` on the same path. `PUT` is the renewing upsert, so it is
  idempotent.
- **Step names are path-escaped**; a name with `/` is accepted percent-encoded.

## Risks / Trade-offs

- A quarantine hides a regression's health signal for up to 30 days. The label
  stays in the list and the expiry is shown, and a step that starts failing
  consistently still raises the failure-rate signal through the run figures.
- A renamed step drops out of quarantine. That fails visibly, since the new
  name starts raising the signal.
- Two reads (steps, then the active set) per summary. The set is tiny and read
  once per request; revisit with the other assumptions in ADR 0004.

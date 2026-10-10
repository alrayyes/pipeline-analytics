# Design

## Decisions

**One mapping, in Go.** `metrics.OutcomeOf(status, conclusion)` is the only
place a forge status becomes a meaning, and it reuses the `isFailed` the
failure insights already use, so "failed" can't mean two things.

**The mapping.**

| Forge state | Outcome |
| --- | --- |
| conclusion `success` | `passed` |
| conclusion `failure` or `timed_out` | `failed` |
| conclusion `cancelled` | `cancelled` |
| conclusion `skipped` | `skipped` |
| no conclusion, status `in_progress` | `running` |
| no conclusion, status `queued`, `waiting`, `pending` or `requested` | `queued` |
| anything else, including a conclusion we don't recognise | `unknown` |

A conclusion wins over a stale `in_progress` status, and an unrecognised
state is `unknown`, never `passed`: a false all-clear is worse than an
unlabelled run.

**`queued` is separate from `running`.** The run list's `running` filter
covers both (anything not completed); the outcome keeps them apart so the
UI can say "Queued" without reading a status string. A client that polls
while work is pending treats both as pending.

**Raw values stay.** `status` and `conclusion` remain, for the run detail
page and for anyone who wants the forge's own words. A `timed_out` step is
`failed` by outcome and still `timed_out` by conclusion.

**The label comes from the outcome.** A timed-out step now reads "Failed"
in the card; the run detail page still shows `timed_out`. Accepted: the
distinction was never decided by anything but the string.

## Risks / Trade-offs

- **New required response field.** Existing consumers that ignore unknown
  fields are unaffected; strict ones need the regenerated SDK. The SDKs
  regenerate automatically on a spec change.
- **A new forge conclusion** shows as `unknown` until the mapping learns it,
  which is visible and safe.

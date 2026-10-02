package metrics

// Outcome is what a run's or step's forge state means, decided here so no
// client interprets status strings itself. It's the API's Outcome schema.
type Outcome string

// The outcomes a run or step can have.
const (
	OutcomePassed    Outcome = "passed"
	OutcomeFailed    Outcome = "failed"
	OutcomeRunning   Outcome = "running"
	OutcomeQueued    Outcome = "queued"
	OutcomeCancelled Outcome = "cancelled"
	OutcomeSkipped   Outcome = "skipped"
	OutcomeUnknown   Outcome = "unknown"
)

// OutcomeOf maps a forge status and conclusion to an Outcome.
//
// A conclusion wins over the status: a run that still says in_progress next
// to a conclusion hasn't resumed, its status just hasn't caught up. A
// failure is whatever isFailed says, the same rule the failure insights use,
// so "failed" can't mean two things. Anything unrecognised is OutcomeUnknown,
// never a pass: a false all-clear is worse than an unlabelled run.
func OutcomeOf(status, conclusion string) Outcome {
	switch {
	case conclusion == "success":
		return OutcomePassed
	case isFailed(conclusion):
		return OutcomeFailed
	case conclusion == "cancelled":
		return OutcomeCancelled
	case conclusion == "skipped":
		return OutcomeSkipped
	case conclusion != "":
		return OutcomeUnknown
	}

	switch status {
	case "in_progress":
		return OutcomeRunning
	case "queued", "waiting", "pending", "requested":
		return OutcomeQueued
	default:
		return OutcomeUnknown
	}
}

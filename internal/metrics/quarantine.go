package metrics

import "time"

// QuarantineDuration is how long a quarantine stays in force from the time it
// was marked or last renewed.
const QuarantineDuration = 30 * 24 * time.Hour

// QuarantineKey names the step a quarantine marks: one step name within one
// pipeline, the same identity FlakyStep uses.
type QuarantineKey struct {
	Pipeline PipelineRef
	Step     string
}

// Quarantine is a person's mark on a flaky step. It mutes the step's
// contribution to its pipeline's health signal and to the flaky-step ratio and
// changes no other figure.
type Quarantine struct {
	Note          string
	QuarantinedAt time.Time
	ExpiresAt     time.Time
}

// countsAsFlaky is the one place that decides whether a flaky step feeds a
// health signal or the flaky-step ratio: a quarantined step is still flaky, it
// just doesn't count.
func countsAsFlaky(step Step) bool {
	return step.Flaky && step.Quarantine == nil
}

// withQuarantines attaches the active mark to each step it names in pipeline.
func withQuarantines(steps []Step, pipeline PipelineRef, active map[QuarantineKey]Quarantine) []Step {
	for i := range steps {
		if q, ok := active[QuarantineKey{Pipeline: pipeline, Step: steps[i].Name}]; ok {
			steps[i].Quarantine = &q
		}
	}

	return steps
}

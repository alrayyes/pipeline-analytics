package metrics_test

import (
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestOutcomeOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     string
		conclusion string
		want       metrics.Outcome
	}{
		{"a successful conclusion passed", "completed", "success", metrics.OutcomePassed},
		{"a failure failed", "completed", "failure", metrics.OutcomeFailed},
		{"a timeout failed", "completed", "timed_out", metrics.OutcomeFailed},
		{"a cancelled conclusion is cancelled", "completed", "cancelled", metrics.OutcomeCancelled},
		{"a skipped conclusion is skipped", "completed", "skipped", metrics.OutcomeSkipped},
		{"in progress with no conclusion is running", "in_progress", "", metrics.OutcomeRunning},
		{"queued with no conclusion is queued", "queued", "", metrics.OutcomeQueued},
		{"waiting is queued", "waiting", "", metrics.OutcomeQueued},
		{"pending is queued", "pending", "", metrics.OutcomeQueued},
		{"requested is queued", "requested", "", metrics.OutcomeQueued},
		{"a conclusion wins over a stale in_progress status: success", "in_progress", "success", metrics.OutcomePassed},
		{"a conclusion wins over a stale in_progress status: failure", "in_progress", "failure", metrics.OutcomeFailed},
		{"completed with no conclusion is unknown, not passed", "completed", "", metrics.OutcomeUnknown},
		{"a conclusion we don't recognise is unknown: neutral", "completed", "neutral", metrics.OutcomeUnknown},
		{"a conclusion we don't recognise is unknown: action_required", "completed", "action_required", metrics.OutcomeUnknown},
		{"a status we don't recognise is unknown", "mystery", "", metrics.OutcomeUnknown},
		{"nothing at all is unknown", "", "", metrics.OutcomeUnknown},
		{"matching is exact: an upper-case conclusion isn't recognised", "completed", "SUCCESS", metrics.OutcomeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, metrics.OutcomeOf(tt.status, tt.conclusion))
		})
	}
}

func TestOutcomeOf_AgreesWithTheFailureInsights(t *testing.T) {
	t.Parallel()

	// "Failed" can't mean two things: every conclusion the failure insights
	// count as a failed run is a failed outcome here.
	for _, conclusion := range []string{"failure", "timed_out"} {
		require.Equal(t, metrics.OutcomeFailed, metrics.OutcomeOf("completed", conclusion), conclusion)
	}
}

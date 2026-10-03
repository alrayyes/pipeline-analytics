package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func listFlaky(t *testing.T, limit, offset int, steps ...metrics.WindowStep) ([]metrics.FlakyStep, bool) {
	t.Helper()

	got, hasMore, err := metrics.NewService(&fakeStore{windowSteps: steps}).ListFlakySteps(
		context.Background(), insightsNow, metrics.InsightWindow{Duration: 24 * time.Hour}, metrics.InsightFilter{}, limit, offset)
	require.NoError(t, err)

	return got, hasMore
}

func flakyNames(steps []metrics.FlakyStep) []string {
	names := make([]string, 0, len(steps))
	for _, s := range steps {
		names = append(names, s.Pipeline.Name+"/"+s.Name)
	}

	return names
}

func TestService_ListFlakySteps(t *testing.T) {
	t.Parallel()

	t.Run("lists only steps that both passed and failed", func(t *testing.T) {
		t.Parallel()

		got, hasMore := listFlaky(t, 0, 0,
			windowStep("CI", "test", "success", 2), windowStep("CI", "test", "failure", 1), // flaky
			windowStep("CI", "build", "success", 2), windowStep("CI", "build", "success", 1),
			windowStep("CI", "lint", "failure", 2), windowStep("CI", "lint", "failure", 1),
		)

		require.Equal(t, []string{"CI/test"}, flakyNames(got))
		require.False(t, hasMore)
	})

	t.Run("reports the flake rate and run count", func(t *testing.T) {
		t.Parallel()

		got, _ := listFlaky(t, 0, 0,
			windowStep("CI", "test", "success", 4), windowStep("CI", "test", "success", 3),
			windowStep("CI", "test", "failure", 2), windowStep("CI", "test", "success", 1),
		)

		require.Len(t, got, 1)
		require.InDelta(t, 0.25, got[0].FlakeRate, 1e-9)
		require.Equal(t, 4, got[0].RunCount)
	})

	t.Run("recent outcomes run oldest first", func(t *testing.T) {
		t.Parallel()

		got, _ := listFlaky(t, 0, 0,
			windowStep("CI", "test", "failure", 1), // newest, listed first on purpose
			windowStep("CI", "test", "success", 3), windowStep("CI", "test", "timed_out", 2),
		)

		require.Len(t, got, 1)
		require.Equal(t, []metrics.Outcome{metrics.OutcomePassed, metrics.OutcomeFailed, metrics.OutcomeFailed}, got[0].RecentOutcomes)
	})

	t.Run("keeps only the 40 most recent outcomes, counting every run", func(t *testing.T) {
		t.Parallel()

		var steps []metrics.WindowStep

		for i := 1; i <= 50; i++ {
			conclusion := "success"
			if i == 1 {
				conclusion = "failure" // the newest run
			}

			steps = append(steps, windowStepAt("CI", "test", conclusion, time.Duration(i)*10*time.Minute))
		}

		got, _ := listFlaky(t, 0, 0, steps...)

		require.Len(t, got, 1)
		require.Equal(t, 50, got[0].RunCount)
		require.Len(t, got[0].RecentOutcomes, 40)
		require.Equal(t, metrics.OutcomeFailed, got[0].RecentOutcomes[39], "the newest run is last")
	})

	t.Run("ranks by flake rate, then run count, then pipeline and step name", func(t *testing.T) {
		t.Parallel()

		got, _ := listFlaky(t, 0, 0,
			windowStep("a", "mild", "success", 3), windowStep("a", "mild", "success", 2), windowStep("a", "mild", "failure", 1), // 1/3
			windowStep("a", "bad", "success", 2), windowStep("a", "bad", "failure", 1), // 1/2
			windowStep("b", "tie", "success", 2), windowStep("b", "tie", "failure", 1), // 1/2
			windowStep("a", "big", "success", 4), windowStep("a", "big", "failure", 3), windowStep("a", "big", "failure", 2), windowStep("a", "big", "success", 1), // 2/4
		)

		require.Equal(t, []string{"a/big", "a/bad", "b/tie", "a/mild"}, flakyNames(got))
	})

	t.Run("the same step name in two pipelines is two entries", func(t *testing.T) {
		t.Parallel()

		got, _ := listFlaky(t, 0, 0,
			windowStep("web", "test", "success", 2), windowStep("web", "test", "failure", 1),
			windowStep("api", "test", "success", 2), windowStep("api", "test", "failure", 1),
		)

		require.Equal(t, []string{"api/test", "web/test"}, flakyNames(got))
	})

	t.Run("pages the ranked list and says whether more follow", func(t *testing.T) {
		t.Parallel()

		steps := []metrics.WindowStep{
			windowStep("a", "one", "success", 2), windowStep("a", "one", "failure", 1),
			windowStep("a", "two", "success", 2), windowStep("a", "two", "failure", 1),
			windowStep("a", "three", "success", 2), windowStep("a", "three", "failure", 1),
		}

		first, hasMore := listFlaky(t, 2, 0, steps...)
		require.Equal(t, []string{"a/one", "a/three"}, flakyNames(first))
		require.True(t, hasMore)

		last, hasMore := listFlaky(t, 2, 2, steps...)
		require.Equal(t, []string{"a/two"}, flakyNames(last))
		require.False(t, hasMore)
	})

	t.Run("no flaky steps is an empty list", func(t *testing.T) {
		t.Parallel()

		got, hasMore := listFlaky(t, 0, 0)

		require.Empty(t, got)
		require.False(t, hasMore)
	})
}

func windowStepAt(pipeline, step, conclusion string, ago time.Duration) metrics.WindowStep {
	ws := windowStep(pipeline, step, conclusion, 0)
	started := insightsNow.Add(-ago)
	ws.Step.RunStartedAt = &started
	ws.Step.RunID = pipeline + started.String()

	return ws
}

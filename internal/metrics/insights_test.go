package metrics_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

var insightsNow = time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)

func TestParseInsightWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want time.Duration
	}{
		{"24h", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"30d", 30 * 24 * time.Hour},
		{"", metrics.DefaultInsightWindow},
		{"12", metrics.DefaultInsightWindow}, // a run count is the per-pipeline window's, not this one's
		{"banana", metrics.DefaultInsightWindow},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.raw), func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, metrics.ParseInsightWindow(tt.raw).Duration)
		})
	}
}

// windowRun is a run that started the given number of hours before
// insightsNow.
func windowRun(pipeline, conclusion string, hoursAgo int) metrics.WindowRun {
	started := insightsNow.Add(-time.Duration(hoursAgo) * time.Hour)
	completed := started.Add(10 * time.Minute)
	status := "completed"

	if conclusion == "" {
		status, completed = "in_progress", time.Time{}
	}

	run := metrics.RunRecord{ID: fmt.Sprintf("%s-%d-%s", pipeline, hoursAgo, conclusion), Status: status, Conclusion: conclusion, StartedAt: &started}
	if !completed.IsZero() {
		run.CompletedAt = &completed
	}

	return metrics.WindowRun{Pipeline: metrics.PipelineRef{RepoID: "repo-1", Name: pipeline}, Run: run}
}

func repeat(n int, build func(i int) metrics.WindowRun) []metrics.WindowRun {
	runs := make([]metrics.WindowRun, 0, n)
	for i := range n {
		runs = append(runs, build(i))
	}

	return runs
}

func TestService_GetFailureInsights(t *testing.T) {
	t.Parallel()

	day := metrics.InsightWindow{Duration: 24 * time.Hour}

	t.Run("counts runs and computes the pass rate over concluded runs only", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{windowRuns: []metrics.WindowRun{
			windowRun("CI", "success", 1), windowRun("CI", "success", 2), windowRun("CI", "success", 3),
			windowRun("CI", "failure", 4),
			windowRun("CI", "cancelled", 5), // neither passed nor failed
			windowRun("CI", "", 1),          // still running
		}}

		got, err := metrics.NewService(store).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{})
		require.NoError(t, err)

		require.Equal(t, 6, got.TotalRuns)
		require.Equal(t, 1, got.FailedRuns)
		require.NotNil(t, got.PassRate)
		require.InDelta(t, 0.75, *got.PassRate, 1e-9) // 3 of 4 concluded
	})

	t.Run("a timed-out run counts as failed", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{windowRuns: []metrics.WindowRun{windowRun("CI", "success", 1), windowRun("CI", "timed_out", 2)}}

		got, err := metrics.NewService(store).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{})
		require.NoError(t, err)
		require.Equal(t, 1, got.FailedRuns)
	})

	t.Run("reports the pass-rate change versus the preceding window in percentage points", func(t *testing.T) {
		t.Parallel()

		// Current 24h: 392 of 500 pass (78.4%). Previous 24h: 413 of 500 (82.6%).
		var runs []metrics.WindowRun

		runs = append(runs, repeat(500, func(i int) metrics.WindowRun {
			return windowRun(fmt.Sprintf("p%d", i), map[bool]string{true: "success", false: "failure"}[i < 392], 1)
		})...)
		runs = append(runs, repeat(500, func(i int) metrics.WindowRun {
			return windowRun(fmt.Sprintf("p%d", i), map[bool]string{true: "success", false: "failure"}[i < 413], 30)
		})...)

		got, err := metrics.NewService(&fakeStore{windowRuns: runs}).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{})
		require.NoError(t, err)

		require.NotNil(t, got.PassRateDelta)
		require.InDelta(t, -4.2, *got.PassRateDelta, 1e-9)
	})

	t.Run("no concluded runs in the window leaves the pass rate absent", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{windowRuns: []metrics.WindowRun{windowRun("CI", "", 1)}}

		got, err := metrics.NewService(store).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{})
		require.NoError(t, err)

		require.Equal(t, 1, got.TotalRuns)
		require.Nil(t, got.PassRate)
		require.Nil(t, got.PassRateDelta)
	})

	t.Run("an empty preceding window leaves the delta absent, not the full rate", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{windowRuns: []metrics.WindowRun{windowRun("CI", "success", 1), windowRun("CI", "failure", 2)}}

		got, err := metrics.NewService(store).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{})
		require.NoError(t, err)

		require.NotNil(t, got.PassRate)
		require.Nil(t, got.PassRateDelta)
	})

	t.Run("ranks pipelines by failed runs, highest first, omitting ones that never failed", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{windowRuns: []metrics.WindowRun{
			windowRun("api", "failure", 1), windowRun("api", "success", 2),
			windowRun("web", "failure", 1), windowRun("web", "failure", 2), windowRun("web", "failure", 3),
			windowRun("docs", "success", 1),
		}}

		got, err := metrics.NewService(store).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{})
		require.NoError(t, err)

		require.Len(t, got.TopFailingPipelines, 2)
		require.Equal(t, "web", got.TopFailingPipelines[0].Pipeline.Name)
		require.Equal(t, 3, got.TopFailingPipelines[0].FailedRuns)
		require.Equal(t, 3, got.TopFailingPipelines[0].Runs)
		require.Equal(t, "api", got.TopFailingPipelines[1].Pipeline.Name)
		require.Equal(t, 2, got.TopFailingPipelines[1].Runs)
	})

	t.Run("scopes to one repo", func(t *testing.T) {
		t.Parallel()

		other := windowRun("CI", "failure", 1)
		other.Pipeline.RepoID = "repo-2"
		store := &fakeStore{windowRuns: []metrics.WindowRun{windowRun("CI", "success", 1), other}}

		got, err := metrics.NewService(store).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{RepoID: "repo-1"})
		require.NoError(t, err)
		require.Equal(t, 1, got.TotalRuns)
		require.Zero(t, got.FailedRuns)
	})
}

// finishedRun is a concluded run that completed the given number of minutes
// before insightsNow, for recovery timing.
func finishedRun(pipeline, conclusion string, completedMinAgo int) metrics.WindowRun {
	completed := insightsNow.Add(-time.Duration(completedMinAgo) * time.Minute)
	started := completed.Add(-5 * time.Minute)

	return metrics.WindowRun{
		Pipeline: metrics.PipelineRef{RepoID: "repo-1", Name: pipeline},
		Run: metrics.RunRecord{
			ID:          fmt.Sprintf("%s-%d-%s", pipeline, completedMinAgo, conclusion),
			Status:      "completed",
			Conclusion:  conclusion,
			StartedAt:   &started,
			CompletedAt: &completed,
		},
	}
}

func TestService_GetFailureInsights_MTTR(t *testing.T) {
	t.Parallel()

	day := metrics.InsightWindow{Duration: 24 * time.Hour}

	mttr := func(t *testing.T, runs ...metrics.WindowRun) *time.Duration {
		t.Helper()

		got, err := metrics.NewService(&fakeStore{windowRuns: runs}).GetFailureInsights(context.Background(), insightsNow, day, metrics.InsightFilter{})
		require.NoError(t, err)

		return got.MTTR
	}

	t.Run("a failure fixed by the next success contributes the gap between them", func(t *testing.T) {
		t.Parallel()

		got := mttr(t, finishedRun("CI", "failure", 100), finishedRun("CI", "success", 60))
		require.NotNil(t, got)
		require.Equal(t, 40*time.Minute, *got)
	})

	t.Run("averages across recoveries, and across pipelines", func(t *testing.T) {
		t.Parallel()

		got := mttr(t,
			finishedRun("CI", "failure", 200), finishedRun("CI", "success", 180), // 20m
			finishedRun("api", "failure", 100), finishedRun("api", "success", 40), // 60m
		)
		require.NotNil(t, got)
		require.Equal(t, 40*time.Minute, *got)
	})

	t.Run("repeated failures count from the first one", func(t *testing.T) {
		t.Parallel()

		got := mttr(t,
			finishedRun("CI", "failure", 120), finishedRun("CI", "failure", 90), finishedRun("CI", "success", 60),
		)
		require.NotNil(t, got)
		require.Equal(t, 60*time.Minute, *got)
	})

	t.Run("a failure that never recovered contributes nothing", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, mttr(t, finishedRun("CI", "success", 100), finishedRun("CI", "failure", 60)))
	})

	t.Run("cancelled and running runs neither break nor end an outage", func(t *testing.T) {
		t.Parallel()

		got := mttr(t,
			finishedRun("CI", "failure", 100), finishedRun("CI", "cancelled", 80), finishedRun("CI", "success", 60),
		)
		require.NotNil(t, got)
		require.Equal(t, 40*time.Minute, *got)
	})

	t.Run("a pipeline that only passed has no MTTR", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, mttr(t, finishedRun("CI", "success", 100), finishedRun("CI", "success", 60)))
	})

	t.Run("runs given newest-first are ordered before pairing", func(t *testing.T) {
		t.Parallel()

		got := mttr(t, finishedRun("CI", "success", 60), finishedRun("CI", "failure", 100))
		require.NotNil(t, got)
		require.Equal(t, 40*time.Minute, *got)
	})
}

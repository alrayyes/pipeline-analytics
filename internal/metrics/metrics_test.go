package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestPipelineRef_IDRoundTrip(t *testing.T) {
	t.Parallel()

	ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}

	got, err := metrics.ParsePipelineID(ref.ID())
	require.NoError(t, err)
	require.Equal(t, ref, got)
}

func TestParsePipelineID_Invalid(t *testing.T) {
	t.Parallel()

	_, err := metrics.ParsePipelineID("not-valid-base64!!!")
	require.ErrorIs(t, err, metrics.ErrInvalidPipelineID)
}

func TestParseWindow(t *testing.T) {
	t.Parallel()

	t.Run("empty falls back to the default run count", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, metrics.Window{RunCount: metrics.DefaultWindowRunCount}, metrics.ParseWindow(""))
	})

	t.Run("a positive integer is a run count", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, metrics.Window{RunCount: 7}, metrics.ParseWindow("7"))
	})

	t.Run("garbage falls back to the default", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, metrics.Window{RunCount: metrics.DefaultWindowRunCount}, metrics.ParseWindow("not-a-number"))
	})
}

// fakeStore is an in-memory metrics.Store fixture: every method reads
// straight from the fields below, set up per test.
type fakeStore struct {
	pipelines []metrics.PipelineRef
	runs      map[metrics.PipelineRef][]metrics.RunRecord
	steps     map[metrics.PipelineRef][]metrics.StepOccurrence
	usage     map[string][]metrics.UsageRecord
}

func (f *fakeStore) ListPipelines(context.Context) ([]metrics.PipelineRef, error) {
	return f.pipelines, nil
}

func (f *fakeStore) PipelineRuns(_ context.Context, ref metrics.PipelineRef, window metrics.Window) ([]metrics.RunRecord, error) {
	runs := f.runs[ref]
	if window.RunCount > 0 && len(runs) > window.RunCount {
		runs = runs[:window.RunCount]
	}

	return runs, nil
}

func (f *fakeStore) PipelineSteps(_ context.Context, ref metrics.PipelineRef, _ metrics.Window) ([]metrics.StepOccurrence, error) {
	return f.steps[ref], nil
}

func (f *fakeStore) RepoUsage(_ context.Context, repoID string, _ metrics.Window) ([]metrics.UsageRecord, error) {
	return f.usage[repoID], nil
}

func t1(offsetMinutes int) *time.Time {
	t := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(offsetMinutes) * time.Minute)

	return &t
}

func completedRun(startMin, durationMin int, conclusion string) metrics.RunRecord {
	return metrics.RunRecord{
		Status:      "completed",
		Conclusion:  conclusion,
		StartedAt:   t1(startMin),
		CompletedAt: t1(startMin + durationMin),
	}
}

func TestService_ListPipelines(t *testing.T) {
	t.Parallel()

	t.Run("a pipeline with a low failure rate and no flaky steps is healthy", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ref},
			runs: map[metrics.PipelineRef][]metrics.RunRecord{
				ref: {
					completedRun(100, 5, "success"),
					completedRun(90, 5, "success"),
					completedRun(80, 5, "success"),
				},
			},
		}

		service := metrics.NewService(store)
		pipelines, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Len(t, pipelines, 1)
		require.Equal(t, metrics.HealthHealthy, pipelines[0].HealthStatus)
		require.Empty(t, pipelines[0].TriggeredSignals)
		require.Equal(t, "repo-1", pipelines[0].RepoID)
		require.Equal(t, "CI", pipelines[0].Name)
	})

	t.Run("an elevated recent failure rate is flagged", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ref},
			runs: map[metrics.PipelineRef][]metrics.RunRecord{
				ref: {
					completedRun(100, 5, "failure"),
					completedRun(90, 5, "failure"),
					completedRun(80, 5, "success"),
				},
			},
		}

		service := metrics.NewService(store)
		pipelines, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Equal(t, metrics.HealthUnhealthy, pipelines[0].HealthStatus)
		require.Contains(t, pipelines[0].TriggeredSignals, metrics.SignalFailureRate)
	})

	t.Run("a flaky step is flagged even when runs all pass", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ref},
			runs: map[metrics.PipelineRef][]metrics.RunRecord{
				ref: {completedRun(100, 5, "success"), completedRun(90, 5, "success")},
			},
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ref: {
					{Name: "flaky-test", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)},
					{Name: "flaky-test", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)},
				},
			},
		}

		service := metrics.NewService(store)
		pipelines, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Equal(t, metrics.HealthUnhealthy, pipelines[0].HealthStatus)
		require.Contains(t, pipelines[0].TriggeredSignals, metrics.SignalFlakyStep)
	})

	t.Run("a significant duration regression is flagged", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}

		// Prior window: five fast, older runs (5-minute duration each).
		var prior []metrics.RunRecord
		for i := range 5 {
			prior = append(prior, completedRun(i*20, 5, "success"))
		}
		// Current window: five slow, more recent runs (20-minute duration
		// each). The store returns runs most-recent-first, so these come
		// before prior in the slice.
		var current []metrics.RunRecord
		for i := range 5 {
			current = append(current, completedRun(200+i*20, 20, "success"))
		}

		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ref},
			runs:      map[metrics.PipelineRef][]metrics.RunRecord{ref: append(current, prior...)},
		}

		service := metrics.NewService(store)
		pipelines, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 5})
		require.NoError(t, err)
		require.Equal(t, metrics.HealthUnhealthy, pipelines[0].HealthStatus)
		require.Contains(t, pipelines[0].TriggeredSignals, metrics.SignalDurationRegression)
	})
}

func TestService_GetPipeline(t *testing.T) {
	t.Parallel()

	t.Run("returns a duration and failure-rate trend with real data points", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			runs: map[metrics.PipelineRef][]metrics.RunRecord{
				ref: {
					completedRun(100, 5, "success"),
					completedRun(90, 5, "failure"),
					completedRun(80, 5, "success"),
				},
			},
		}

		service := metrics.NewService(store)
		detail, err := service.GetPipeline(context.Background(), ref.ID(), metrics.Window{RunCount: 10})
		require.NoError(t, err)

		require.NotEmpty(t, detail.DurationTrend.Timestamps)
		require.NotEmpty(t, detail.DurationTrend.P50)
		require.NotEmpty(t, detail.DurationTrend.P90)
		require.NotEmpty(t, detail.FailureRateTrend.Rate)
	})

	t.Run("a pipeline with no run history is not found", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "ghost"}
		service := metrics.NewService(&fakeStore{})

		_, err := service.GetPipeline(context.Background(), ref.ID(), metrics.Window{RunCount: 10})
		require.ErrorIs(t, err, metrics.ErrPipelineNotFound)
	})

	t.Run("an invalid pipeline id is rejected", func(t *testing.T) {
		t.Parallel()

		service := metrics.NewService(&fakeStore{})

		_, err := service.GetPipeline(context.Background(), "not-valid", metrics.Window{RunCount: 10})
		require.ErrorIs(t, err, metrics.ErrInvalidPipelineID)
	})
}

func TestService_GetPipelineSteps(t *testing.T) {
	t.Parallel()

	t.Run("ranks steps by duration contribution, highest first", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ref: {
					{Name: "fast", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)},
					{Name: "slow", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(10)},
				},
			},
		}

		service := metrics.NewService(store)
		steps, err := service.GetPipelineSteps(context.Background(), ref.ID(), metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Len(t, steps, 2)
		require.Equal(t, "slow", steps[0].Name)
		require.Equal(t, "fast", steps[1].Name)
		require.Greater(t, steps[0].DurationContributionSeconds, steps[1].DurationContributionSeconds)
	})

	t.Run("separates queue time (the job's wait) from execution time (the step's own run)", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ref: {
					{
						Name: "build", Status: "completed", Conclusion: "success",
						JobQueuedAt: t1(0), JobStartedAt: t1(30), // 30 minutes queued
						StartedAt: t1(30), CompletedAt: t1(35), // 5 minutes executing
					},
				},
			},
		}

		service := metrics.NewService(store)
		steps, err := service.GetPipelineSteps(context.Background(), ref.ID(), metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Len(t, steps, 1)
		require.InDelta(t, 30*60, steps[0].QueueSeconds, 0.01)
		require.InDelta(t, 5*60, steps[0].ExecSeconds, 0.01)
	})

	t.Run("a step that alternates pass and fail is flaky, distinct from one that always fails", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ref: {
					{Name: "flaky", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)},
					{Name: "flaky", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)},
					{Name: "broken", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)},
					{Name: "broken", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)},
				},
			},
		}

		service := metrics.NewService(store)
		steps, err := service.GetPipelineSteps(context.Background(), ref.ID(), metrics.Window{RunCount: 10})
		require.NoError(t, err)

		byName := map[string]metrics.Step{}
		for _, s := range steps {
			byName[s.Name] = s
		}

		require.True(t, byName["flaky"].Flaky)
		require.False(t, byName["broken"].Flaky)
		require.InDelta(t, 1.0, byName["broken"].FailureRate, 0.001)
		require.InDelta(t, 0.5, byName["flaky"].FailureRate, 0.001)
	})
}

func TestService_GetRepoUsage(t *testing.T) {
	t.Parallel()

	t.Run("reports runner minutes broken down by workflow, highest first", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			usage: map[string][]metrics.UsageRecord{
				"repo-1": {
					{PipelineName: "CI", ExecSeconds: 120},
					{PipelineName: "CI", ExecSeconds: 60},
					{PipelineName: "Deploy", ExecSeconds: 600},
				},
			},
		}

		service := metrics.NewService(store)
		usage, err := service.GetRepoUsage(context.Background(), "repo-1", metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Len(t, usage, 2)
		require.Equal(t, "Deploy", usage[0].Workflow)
		require.InDelta(t, 10.0, usage[0].RunnerMinutes, 0.001)
		require.Equal(t, "CI", usage[1].Workflow)
		require.InDelta(t, 3.0, usage[1].RunnerMinutes, 0.001)
	})
}

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
	runSteps  map[string][]metrics.StepOccurrence
	usage     map[string][]metrics.UsageRecord
}

// ListPipelines applies filter.RepoID and pagination the same way the real
// store does, so a Service-level test can verify the filter reaches
// through. filter.Forge is ignored -- the fake has no notion of a repo's
// forge, and that filter is covered at the sqlite store layer, where the
// join it needs actually exists.
func (f *fakeStore) ListPipelines(_ context.Context, filter metrics.PipelineListFilter) ([]metrics.PipelineRef, bool, error) {
	var matching []metrics.PipelineRef

	for _, ref := range f.pipelines {
		if filter.RepoID != "" && ref.RepoID != filter.RepoID {
			continue
		}

		matching = append(matching, ref)
	}

	if filter.Limit <= 0 {
		return matching, false, nil
	}

	offset := min(filter.Offset, len(matching))
	end := min(offset+filter.Limit, len(matching))
	hasMore := len(matching) > offset+filter.Limit

	return matching[offset:end], hasMore, nil
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

func (f *fakeStore) RunSteps(_ context.Context, runID string) ([]metrics.StepOccurrence, error) {
	return f.runSteps[runID], nil
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
		pipelines, _, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10}, metrics.PipelineListFilter{})
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
		pipelines, _, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10}, metrics.PipelineListFilter{})
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
		pipelines, _, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10}, metrics.PipelineListFilter{})
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
		pipelines, _, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 5}, metrics.PipelineListFilter{})
		require.NoError(t, err)
		require.Equal(t, metrics.HealthUnhealthy, pipelines[0].HealthStatus)
		require.Contains(t, pipelines[0].TriggeredSignals, metrics.SignalDurationRegression)
	})

	t.Run("LastRunAt is the most recent run's start time", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ref},
			runs: map[metrics.PipelineRef][]metrics.RunRecord{
				// Most-recent-first, per the Store contract.
				ref: {
					completedRun(100, 5, "success"),
					completedRun(90, 5, "success"),
				},
			},
		}

		service := metrics.NewService(store)
		pipelines, _, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10}, metrics.PipelineListFilter{})
		require.NoError(t, err)
		require.Equal(t, t1(100), pipelines[0].LastRunAt)
	})

	t.Run("the filter and its hasMore reach through to the store", func(t *testing.T) {
		t.Parallel()

		wanted := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		other := metrics.PipelineRef{RepoID: "repo-2", Name: "Deploy"}
		store := &fakeStore{
			pipelines: []metrics.PipelineRef{wanted, other},
			runs: map[metrics.PipelineRef][]metrics.RunRecord{
				wanted: {completedRun(100, 5, "success")},
				other:  {completedRun(100, 5, "success")},
			},
		}

		service := metrics.NewService(store)
		pipelines, hasMore, err := service.ListPipelines(context.Background(), metrics.Window{RunCount: 10}, metrics.PipelineListFilter{RepoID: "repo-1", Limit: 1})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, pipelines, 1)
		require.Equal(t, "repo-1", pipelines[0].RepoID)
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
		require.Equal(t, 2, byName["broken"].FailureCount)
		require.Equal(t, 1, byName["flaky"].FailureCount)
	})
}

func TestService_ListFlakyRuns(t *testing.T) {
	t.Parallel()

	t.Run("only the step's failed occurrences show, most-recent-first", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ref: {
					{Name: "flaky", Conclusion: "success", RunID: "run-1", RunStartedAt: t1(0), JobForgeURL: "https://forge/run-1"},
					{Name: "flaky", Conclusion: "failure", RunID: "run-2", RunStartedAt: t1(10), JobForgeURL: "https://forge/run-2"},
					{Name: "flaky", Conclusion: "failure", RunID: "run-3", RunStartedAt: t1(20), JobForgeURL: "https://forge/run-3"},
					{Name: "other-step", Conclusion: "failure", RunID: "run-4", RunStartedAt: t1(30), JobForgeURL: "https://forge/run-4"},
				},
			},
		}

		service := metrics.NewService(store)
		runs, err := service.ListFlakyRuns(context.Background(), ref.ID(), "flaky", metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Len(t, runs, 2)
		require.Equal(t, "run-3", runs[0].RunID) // most recent first
		require.Equal(t, "run-2", runs[1].RunID)
		require.Equal(t, "https://forge/run-3", runs[0].ForgeURL)
	})

	t.Run("a step with no failures in the window reports an empty list", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ref: {{Name: "solid", Conclusion: "success", RunID: "run-1", RunStartedAt: t1(0)}},
			},
		}

		service := metrics.NewService(store)
		runs, err := service.ListFlakyRuns(context.Background(), ref.ID(), "solid", metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Empty(t, runs)
	})

	t.Run("an invalid pipeline id is rejected", func(t *testing.T) {
		t.Parallel()

		service := metrics.NewService(&fakeStore{})
		_, err := service.ListFlakyRuns(context.Background(), "not-valid-base64!!!", "solid", metrics.Window{RunCount: 10})
		require.ErrorIs(t, err, metrics.ErrInvalidPipelineID)
	})
}

func TestService_GetRunSteps(t *testing.T) {
	t.Parallel()

	t.Run("returns the run's own steps in recorded order", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{
			runSteps: map[string][]metrics.StepOccurrence{
				"run-1": {
					{Name: "checkout", Status: "completed", Conclusion: "success", RunID: "run-1", RunStartedAt: t1(0), JobForgeURL: "https://forge/run-1/job/1"},
					{Name: "test", Status: "completed", Conclusion: "failure", RunID: "run-1", RunStartedAt: t1(0), JobForgeURL: "https://forge/run-1/job/2"},
				},
			},
		}

		service := metrics.NewService(store)
		detail, err := service.GetRunSteps(context.Background(), "run-1")
		require.NoError(t, err)
		require.Equal(t, "run-1", detail.RunID)
		require.Len(t, detail.Steps, 2)
		require.Equal(t, "checkout", detail.Steps[0].Name)
		require.Equal(t, "test", detail.Steps[1].Name)
		require.Equal(t, "failure", detail.Steps[1].Conclusion)
		require.Equal(t, "https://forge/run-1/job/2", detail.Steps[1].ForgeURL)
	})

	t.Run("an unknown run is not found", func(t *testing.T) {
		t.Parallel()

		service := metrics.NewService(&fakeStore{})
		_, err := service.GetRunSteps(context.Background(), "ghost")
		require.ErrorIs(t, err, metrics.ErrRunNotFound)
	})
}

func TestService_ListUnhealthySteps(t *testing.T) {
	t.Parallel()

	t.Run("only flaky or failing steps show, grouped by pipeline", func(t *testing.T) {
		t.Parallel()

		flakyPipeline := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		healthyPipeline := metrics.PipelineRef{RepoID: "repo-1", Name: "Lint"}
		failingPipeline := metrics.PipelineRef{RepoID: "repo-2", Name: "Deploy"}

		store := &fakeStore{
			pipelines: []metrics.PipelineRef{flakyPipeline, healthyPipeline, failingPipeline},
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				flakyPipeline: {
					{Name: "flaky-test", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)},
					{Name: "flaky-test", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)},
					{Name: "solid-test", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)},
				},
				healthyPipeline: {
					{Name: "lint", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)},
				},
				failingPipeline: {
					{Name: "push-image", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)},
				},
			},
		}

		service := metrics.NewService(store)
		groups, hasMore, err := service.ListUnhealthySteps(context.Background(), metrics.Window{RunCount: 10}, 0, 0)
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, groups, 2)

		byPipeline := map[string]metrics.PipelineStepsGroup{}
		for _, g := range groups {
			byPipeline[g.PipelineName] = g
		}

		_, healthyPresent := byPipeline["Lint"]
		require.False(t, healthyPresent)

		ciGroup := byPipeline["CI"]
		require.Equal(t, "repo-1", ciGroup.RepoID)
		require.Len(t, ciGroup.Steps, 1)
		require.Equal(t, "flaky-test", ciGroup.Steps[0].Name)
		require.True(t, ciGroup.Steps[0].Flaky)

		deployGroup := byPipeline["Deploy"]
		require.Equal(t, "repo-2", deployGroup.RepoID)
		require.Len(t, deployGroup.Steps, 1)
		require.Equal(t, "push-image", deployGroup.Steps[0].Name)
		require.InDelta(t, 1.0, deployGroup.Steps[0].FailureRate, 0.001)
	})

	t.Run("no unhealthy steps anywhere reports an empty list", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ref},
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ref: {{Name: "lint", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)}},
			},
		}

		service := metrics.NewService(store)
		groups, hasMore, err := service.ListUnhealthySteps(context.Background(), metrics.Window{RunCount: 10}, 0, 0)
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Empty(t, groups)
	})

	t.Run("paginates over pipeline groups with an unhealthy step, not every candidate pipeline", func(t *testing.T) {
		t.Parallel()

		// Ordered (repo_id, pipeline_name) the way ListPipelines returns them:
		// repo-1/CI and repo-1/Deploy are unhealthy, repo-1/Lint is healthy and
		// contributes nothing -- pagination has to skip over it rather than
		// counting it as a page slot, since it never becomes a group.
		ciRef := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		lintRef := metrics.PipelineRef{RepoID: "repo-1", Name: "Lint"}
		deployRef := metrics.PipelineRef{RepoID: "repo-1", Name: "Deploy"}
		releaseRef := metrics.PipelineRef{RepoID: "repo-1", Name: "Release"}

		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ciRef, lintRef, deployRef, releaseRef},
			steps: map[metrics.PipelineRef][]metrics.StepOccurrence{
				ciRef:      {{Name: "test", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)}},
				lintRef:    {{Name: "lint", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)}},
				deployRef:  {{Name: "push", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)}},
				releaseRef: {{Name: "tag", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)}},
			},
		}

		service := metrics.NewService(store)

		firstPage, hasMore, err := service.ListUnhealthySteps(context.Background(), metrics.Window{RunCount: 10}, 2, 0)
		require.NoError(t, err)
		require.True(t, hasMore)
		require.Len(t, firstPage, 2)
		require.Equal(t, "CI", firstPage[0].PipelineName)
		require.Equal(t, "Deploy", firstPage[1].PipelineName)

		secondPage, hasMore, err := service.ListUnhealthySteps(context.Background(), metrics.Window{RunCount: 10}, 2, 2)
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, secondPage, 1)
		require.Equal(t, "Release", secondPage[0].PipelineName)
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

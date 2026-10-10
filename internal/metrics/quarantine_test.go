package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

// quarantinedTest is a mark on the "test" step of the "CI" pipeline in repo-1,
// the pair the helpers below make flaky.
func quarantinedTest() map[metrics.QuarantineKey]metrics.Quarantine {
	return map[metrics.QuarantineKey]metrics.Quarantine{
		{Pipeline: metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}, Step: "test"}: {
			Note:          "waits on the shared database",
			QuarantinedAt: insightsNow.Add(-48 * time.Hour),
			ExpiresAt:     insightsNow.Add(28 * 24 * time.Hour),
		},
	}
}

func flakyOccurrences() []metrics.StepOccurrence {
	return []metrics.StepOccurrence{
		{Name: "test", Status: "completed", Conclusion: "success", StartedAt: t1(0), CompletedAt: t1(1)},
		{Name: "test", Status: "completed", Conclusion: "failure", StartedAt: t1(0), CompletedAt: t1(1)},
	}
}

func listWithQuarantine(t *testing.T, runs []metrics.RunRecord, marks map[metrics.QuarantineKey]metrics.Quarantine) metrics.Pipeline {
	t.Helper()

	ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
	store := &fakeStore{
		pipelines:   []metrics.PipelineRef{ref},
		runs:        map[metrics.PipelineRef][]metrics.RunRecord{ref: runs},
		steps:       map[metrics.PipelineRef][]metrics.StepOccurrence{ref: flakyOccurrences()},
		quarantines: marks,
	}

	pipelines, _, err := metrics.NewService(store).ListPipelines(context.Background(), metrics.Window{RunCount: 10}, metrics.PipelineListFilter{})
	require.NoError(t, err)
	require.Len(t, pipelines, 1)

	return pipelines[0]
}

func TestService_Quarantine_Health(t *testing.T) {
	t.Parallel()

	passing := []metrics.RunRecord{completedRun(100, 5, "success"), completedRun(90, 5, "success")}

	t.Run("a flaky step that is not quarantined makes the pipeline unhealthy", func(t *testing.T) {
		t.Parallel()

		got := listWithQuarantine(t, passing, nil)
		require.Equal(t, metrics.HealthUnhealthy, got.HealthStatus)
		require.Contains(t, got.TriggeredSignals, metrics.SignalFlakyStep)
	})

	t.Run("a quarantined flaky step no longer raises the signal", func(t *testing.T) {
		t.Parallel()

		got := listWithQuarantine(t, passing, quarantinedTest())
		require.Equal(t, metrics.HealthHealthy, got.HealthStatus)
		require.NotContains(t, got.TriggeredSignals, metrics.SignalFlakyStep)
	})

	t.Run("failures still count: the failure-rate signal survives a quarantine", func(t *testing.T) {
		t.Parallel()

		failing := []metrics.RunRecord{completedRun(100, 5, "failure"), completedRun(90, 5, "failure")}

		got := listWithQuarantine(t, failing, quarantinedTest())
		require.Equal(t, metrics.HealthUnhealthy, got.HealthStatus)
		require.Contains(t, got.TriggeredSignals, metrics.SignalFailureRate)
		require.NotContains(t, got.TriggeredSignals, metrics.SignalFlakyStep)
	})

	t.Run("a mark on another pipeline's step of the same name does nothing", func(t *testing.T) {
		t.Parallel()

		other := map[metrics.QuarantineKey]metrics.Quarantine{
			{Pipeline: metrics.PipelineRef{RepoID: "repo-1", Name: "Deploy"}, Step: "test"}: {ExpiresAt: insightsNow.Add(time.Hour)},
		}

		require.Contains(t, listWithQuarantine(t, passing, other).TriggeredSignals, metrics.SignalFlakyStep)
	})
}

func TestService_Quarantine_Steps(t *testing.T) {
	t.Parallel()

	t.Run("the pipeline's step list carries the mark and keeps the figures", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			pipelines:   []metrics.PipelineRef{ref},
			steps:       map[metrics.PipelineRef][]metrics.StepOccurrence{ref: flakyOccurrences()},
			quarantines: quarantinedTest(),
		}

		steps, err := metrics.NewService(store).GetPipelineSteps(context.Background(), ref.ID(), metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Len(t, steps, 1)
		require.True(t, steps[0].Flaky)
		require.InDelta(t, 0.5, steps[0].FailureRate, 1e-9)
		require.NotNil(t, steps[0].Quarantine)
		require.Equal(t, "waits on the shared database", steps[0].Quarantine.Note)
	})

	t.Run("the unhealthy-steps overview leaves out a quarantined flaky step", func(t *testing.T) {
		t.Parallel()

		ref := metrics.PipelineRef{RepoID: "repo-1", Name: "CI"}
		store := &fakeStore{
			pipelines: []metrics.PipelineRef{ref},
			steps:     map[metrics.PipelineRef][]metrics.StepOccurrence{ref: flakyOccurrences()},
		}

		groups, _, err := metrics.NewService(store).ListUnhealthySteps(context.Background(), metrics.Window{RunCount: 10}, 0, 0)
		require.NoError(t, err)
		require.Len(t, groups, 1)

		store.quarantines = quarantinedTest()

		groups, _, err = metrics.NewService(store).ListUnhealthySteps(context.Background(), metrics.Window{RunCount: 10}, 0, 0)
		require.NoError(t, err)
		require.Empty(t, groups)
	})
}

func TestService_Quarantine_FlakyList(t *testing.T) {
	t.Parallel()

	steps := []metrics.WindowStep{
		windowStep("CI", "test", "success", 2), windowStep("CI", "test", "failure", 1),
	}

	plain, _ := listFlaky(t, 0, 0, steps...)
	require.Len(t, plain, 1)
	require.Nil(t, plain[0].Quarantine)

	got, _, err := metrics.NewService(&fakeStore{windowSteps: steps, quarantines: quarantinedTest()}).ListFlakySteps(
		context.Background(), insightsNow, metrics.InsightWindow{Duration: 24 * time.Hour}, metrics.InsightFilter{}, 0, 0)
	require.NoError(t, err)
	require.Len(t, got, 1)

	t.Run("a quarantined step stays listed with every figure unchanged", func(t *testing.T) {
		require.Equal(t, plain[0].FlakeRate, got[0].FlakeRate)
		require.Equal(t, plain[0].RunCount, got[0].RunCount)
		require.Equal(t, plain[0].RecentOutcomes, got[0].RecentOutcomes)
	})

	t.Run("and reports the mark", func(t *testing.T) {
		require.NotNil(t, got[0].Quarantine)
		require.Equal(t, "waits on the shared database", got[0].Quarantine.Note)
		require.Equal(t, insightsNow.Add(28*24*time.Hour), got[0].Quarantine.ExpiresAt)
	})
}

func TestService_Quarantine_FlakyStepRatio(t *testing.T) {
	t.Parallel()

	steps := []metrics.WindowStep{
		windowStep("CI", "test", "success", 1), windowStep("CI", "test", "failure", 2), // flaky
		windowStep("CI", "build", "success", 1), windowStep("CI", "build", "success", 2),
		windowStep("CI", "lint", "failure", 1), windowStep("CI", "lint", "failure", 2),
		windowStep("CI", "docs", "success", 1),
	}

	insights := func(marks map[metrics.QuarantineKey]metrics.Quarantine) metrics.FailureInsights {
		got, err := metrics.NewService(&fakeStore{windowSteps: steps, quarantines: marks}).GetFailureInsights(
			context.Background(), insightsNow, metrics.InsightWindow{Duration: 24 * time.Hour}, metrics.InsightFilter{})
		require.NoError(t, err)

		return got
	}

	require.InDelta(t, 0.25, insights(nil).FlakyStepRatio, 1e-9)

	// The step still counts among the four; only its flakiness stops counting.
	require.Zero(t, insights(quarantinedTest()).FlakyStepRatio)
}

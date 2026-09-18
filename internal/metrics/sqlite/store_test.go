package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	metricssqlite "github.com/alrayyes/pipeline-analytics/internal/metrics/sqlite"
	"github.com/stretchr/testify/require"
)

// testFixture wires a metrics.Store against the same database an
// ingestion.RunStore writes to, so tests seed data through the real
// ingestion write path rather than hand-crafting rows.
type testFixture struct {
	metrics *metricssqlite.Store
	runs    ingestion.RunStore
	repo    ingestion.Repo
}

func newFixture(t *testing.T) testFixture {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	ingestionStore := ingestionsqlite.NewStore(conn, make([]byte, 32))

	repo, err := ingestionStore.CreateRepo(context.Background(), ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_test",
	})
	require.NoError(t, err)

	return testFixture{
		metrics: metricssqlite.NewStore(conn),
		runs:    ingestionStore,
		repo:    repo,
	}
}

func at(minutesFromEpoch int) *time.Time {
	t := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(minutesFromEpoch) * time.Minute)

	return &t
}

func (f testFixture) seedRun(t *testing.T, pipelineName, forgeRunID string, startedMin, durationMin int, conclusion string) ingestion.Run {
	t.Helper()

	run, err := f.runs.UpsertRun(context.Background(), ingestion.Run{
		RepoID:       f.repo.ID,
		ForgeRunID:   forgeRunID,
		PipelineName: pipelineName,
		Status:       "completed",
		Conclusion:   conclusion,
		StartedAt:    at(startedMin),
		CompletedAt:  at(startedMin + durationMin),
		ForgeURL:     "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + forgeRunID,
	})
	require.NoError(t, err)

	return run
}

func (f testFixture) seedJobWithStep(t *testing.T, runID, forgeJobID, stepName string, queuedMin, startedMin, completedMin int, conclusion string) {
	t.Helper()

	job, err := f.runs.UpsertJob(context.Background(), ingestion.Job{
		RunID:       runID,
		ForgeJobID:  forgeJobID,
		Name:        "build",
		Status:      "completed",
		Conclusion:  conclusion,
		QueuedAt:    at(queuedMin),
		StartedAt:   at(startedMin),
		CompletedAt: at(completedMin),
		ForgeURL:    "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + runID + "/job/" + forgeJobID,
	})
	require.NoError(t, err)

	err = f.runs.ReplaceSteps(context.Background(), job.ID, []ingestion.Step{
		{Number: 1, Name: stepName, Status: "completed", Conclusion: conclusion, StartedAt: at(startedMin), CompletedAt: at(completedMin)},
	})
	require.NoError(t, err)
}

func TestStore_ListPipelines(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.seedRun(t, "CI", "1", 0, 5, "success")
	f.seedRun(t, "Deploy", "2", 10, 5, "success")
	f.seedRun(t, "CI", "3", 20, 5, "success") // same pipeline again -- must not duplicate

	pipelines, err := f.metrics.ListPipelines(context.Background())
	require.NoError(t, err)
	require.Len(t, pipelines, 2)

	names := map[string]bool{}
	for _, p := range pipelines {
		require.Equal(t, f.repo.ID, p.RepoID)

		names[p.Name] = true
	}

	require.True(t, names["CI"])
	require.True(t, names["Deploy"])
}

func TestStore_PipelineRuns(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.seedRun(t, "CI", "1", 0, 5, "success")
	f.seedRun(t, "CI", "2", 10, 5, "failure")
	f.seedRun(t, "CI", "3", 20, 5, "success")

	ref := metrics.PipelineRef{RepoID: f.repo.ID, Name: "CI"}

	t.Run("returns runs most-recent-first", func(t *testing.T) {
		t.Parallel()

		runs, err := f.metrics.PipelineRuns(context.Background(), ref, metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Len(t, runs, 3)
		require.Equal(t, "success", runs[0].Conclusion) // run 3, most recent
		require.Equal(t, "failure", runs[1].Conclusion) // run 2
		require.Equal(t, "success", runs[2].Conclusion) // run 1, oldest
	})

	t.Run("respects the window's run count", func(t *testing.T) {
		t.Parallel()

		runs, err := f.metrics.PipelineRuns(context.Background(), ref, metrics.Window{RunCount: 1})
		require.NoError(t, err)
		require.Len(t, runs, 1)
	})

	t.Run("a pipeline with no runs returns an empty slice", func(t *testing.T) {
		t.Parallel()

		runs, err := f.metrics.PipelineRuns(context.Background(), metrics.PipelineRef{RepoID: f.repo.ID, Name: "ghost"}, metrics.Window{RunCount: 10})
		require.NoError(t, err)
		require.Empty(t, runs)
	})
}

func TestStore_PipelineSteps(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	run := f.seedRun(t, "CI", "1", 0, 30, "success")
	f.seedJobWithStep(t, run.ID, "100", "test", 0, 10, 25, "success")

	ref := metrics.PipelineRef{RepoID: f.repo.ID, Name: "CI"}

	steps, err := f.metrics.PipelineSteps(context.Background(), ref, metrics.Window{RunCount: 10})
	require.NoError(t, err)
	require.Len(t, steps, 1)

	require.Equal(t, "test", steps[0].Name)
	require.Equal(t, "completed", steps[0].Status)
	require.Equal(t, "success", steps[0].Conclusion)
	require.NotNil(t, steps[0].JobQueuedAt)
	require.NotNil(t, steps[0].JobStartedAt)
	require.Contains(t, steps[0].JobForgeURL, "/job/100")
	require.Equal(t, run.ID, steps[0].RunID)
	require.NotNil(t, steps[0].RunStartedAt)
}

func TestStore_RunSteps(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	run := f.seedRun(t, "CI", "1", 0, 30, "failure")
	f.seedJobWithStep(t, run.ID, "100", "test", 0, 10, 25, "failure")

	t.Run("returns the run's own steps, tagged with the run they belong to", func(t *testing.T) {
		t.Parallel()

		steps, err := f.metrics.RunSteps(context.Background(), run.ID)
		require.NoError(t, err)
		require.Len(t, steps, 1)
		require.Equal(t, "test", steps[0].Name)
		require.Equal(t, "failure", steps[0].Conclusion)
		require.Equal(t, run.ID, steps[0].RunID)
		require.NotNil(t, steps[0].RunStartedAt)
		require.Contains(t, steps[0].JobForgeURL, "/job/100")
	})

	t.Run("an unknown run returns an empty slice", func(t *testing.T) {
		t.Parallel()

		steps, err := f.metrics.RunSteps(context.Background(), "ghost")
		require.NoError(t, err)
		require.Empty(t, steps)
	})
}

func TestStore_RepoUsage(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ciRun := f.seedRun(t, "CI", "1", 0, 10, "success")
	deployRun := f.seedRun(t, "Deploy", "2", 20, 10, "success")

	f.seedJobWithStep(t, ciRun.ID, "100", "test", 0, 0, 5, "success")          // 5 minutes exec
	f.seedJobWithStep(t, deployRun.ID, "200", "deploy", 20, 20, 30, "success") // 10 minutes exec

	usage, err := f.metrics.RepoUsage(context.Background(), f.repo.ID, metrics.Window{RunCount: 10})
	require.NoError(t, err)
	require.Len(t, usage, 2)

	byPipeline := map[string]float64{}
	for _, u := range usage {
		byPipeline[u.PipelineName] = u.ExecSeconds
	}

	require.InDelta(t, 5*60, byPipeline["CI"], 0.01)
	require.InDelta(t, 10*60, byPipeline["Deploy"], 0.01)
}

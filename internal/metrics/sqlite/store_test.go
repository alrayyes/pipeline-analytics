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
	runs    *ingestionsqlite.Store
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
	f.seedRun(t, "Deploy", "1", 0, 5, "success")
	f.seedRun(t, "CI", "2", 10, 5, "success")
	f.seedRun(t, "CI", "3", 20, 5, "success") // same pipeline again -- must not duplicate

	pipelines, hasMore, err := f.metrics.ListPipelines(context.Background(), metrics.PipelineListFilter{})
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Len(t, pipelines, 2)

	// Ordered by (repo_id, pipeline_name), not insertion order.
	require.Equal(t, "CI", pipelines[0].Name)
	require.Equal(t, "Deploy", pipelines[1].Name)
	require.Equal(t, f.repo.ID, pipelines[0].RepoID)
}

func TestStore_ListPipelinesPagination(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.seedRun(t, "Build", "1", 0, 5, "success")
	f.seedRun(t, "CI", "2", 10, 5, "success")
	f.seedRun(t, "Deploy", "3", 20, 5, "success")

	t.Run("limit trims the page and reports more remain", func(t *testing.T) {
		t.Parallel()

		pipelines, hasMore, err := f.metrics.ListPipelines(context.Background(), metrics.PipelineListFilter{Limit: 2})
		require.NoError(t, err)
		require.True(t, hasMore)
		require.Len(t, pipelines, 2)
		require.Equal(t, "Build", pipelines[0].Name)
		require.Equal(t, "CI", pipelines[1].Name)
	})

	t.Run("offset returns the next page", func(t *testing.T) {
		t.Parallel()

		pipelines, hasMore, err := f.metrics.ListPipelines(context.Background(), metrics.PipelineListFilter{Limit: 2, Offset: 2})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, pipelines, 1)
		require.Equal(t, "Deploy", pipelines[0].Name)
	})

	t.Run("no limit returns every pipeline unpaginated", func(t *testing.T) {
		t.Parallel()

		pipelines, hasMore, err := f.metrics.ListPipelines(context.Background(), metrics.PipelineListFilter{})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, pipelines, 3)
	})
}

func TestStore_ListPipelinesFilters(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.seedRun(t, "CI", "1", 0, 5, "success")

	otherRepo, err := f.runs.CreateRepo(context.Background(), ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
		Forge:              ingestion.ForgeForgejo,
		Identifier:         "alrayyes/dotfiles",
		ForgejoInstanceURL: "https://git.higherlearning.eu",
		Token:              "forgejo-token-5678",
	})
	require.NoError(t, err)

	otherFixture := testFixture{metrics: f.metrics, runs: f.runs, repo: otherRepo}
	otherFixture.seedRun(t, "Sync", "1", 0, 5, "success")

	t.Run("repoId restricts the list to one tracked repo", func(t *testing.T) {
		t.Parallel()

		pipelines, hasMore, err := f.metrics.ListPipelines(context.Background(), metrics.PipelineListFilter{RepoID: f.repo.ID})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, pipelines, 1)
		require.Equal(t, "CI", pipelines[0].Name)
	})

	t.Run("forge restricts the list via the owning repo", func(t *testing.T) {
		t.Parallel()

		pipelines, hasMore, err := f.metrics.ListPipelines(context.Background(), metrics.PipelineListFilter{Forge: string(ingestion.ForgeForgejo)})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, pipelines, 1)
		require.Equal(t, "Sync", pipelines[0].Name)
	})
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

func TestStore_WindowRuns(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	// Minutes from the epoch: 100 and 200 are inside [150-100, 250), 300 is not.
	f.seedRun(t, "CI", "1", 40, 5, "success")  // before the window
	f.seedRun(t, "CI", "2", 100, 5, "failure") // inside
	f.seedRun(t, "Lint", "3", 200, 5, "success")
	f.seedRun(t, "CI", "4", 300, 5, "success") // after

	t.Run("returns runs started within [since, until) with their pipeline", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowRuns(context.Background(), metrics.RunWindowFilter{Since: *at(50), Until: *at(250)})
		require.NoError(t, err)
		require.Len(t, got, 2)

		names := []string{got[0].Pipeline.Name, got[1].Pipeline.Name}
		require.ElementsMatch(t, []string{"CI", "Lint"}, names)
		require.Equal(t, f.repo.ID, got[0].Pipeline.RepoID)
	})

	t.Run("the upper bound is exclusive", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowRuns(context.Background(), metrics.RunWindowFilter{Since: *at(50), Until: *at(200)})
		require.NoError(t, err)
		require.Len(t, got, 1)
	})

	t.Run("filters by repo", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowRuns(context.Background(), metrics.RunWindowFilter{RepoID: "no-such-repo", Since: *at(0), Until: *at(1000)})
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("filters by forge", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowRuns(context.Background(), metrics.RunWindowFilter{Forge: "forgejo", Since: *at(0), Until: *at(1000)})
		require.NoError(t, err)
		require.Empty(t, got)

		got, err = f.metrics.WindowRuns(context.Background(), metrics.RunWindowFilter{Forge: "github", Since: *at(0), Until: *at(1000)})
		require.NoError(t, err)
		require.Len(t, got, 4)
	})
}

func TestStore_WindowSteps(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	before := f.seedRun(t, "CI", "1", 40, 5, "success")
	inside := f.seedRun(t, "CI", "2", 100, 5, "failure")
	other := f.seedRun(t, "Lint", "3", 200, 5, "success")

	f.seedJobWithStep(t, before.ID, "10", "build", 40, 41, 44, "success")
	f.seedJobWithStep(t, inside.ID, "11", "test", 100, 101, 104, "failure")
	f.seedJobWithStep(t, other.ID, "12", "golangci-lint", 200, 201, 204, "success")

	t.Run("returns steps of runs started in [since, until) with their pipeline", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowSteps(context.Background(), metrics.RunWindowFilter{Since: *at(50), Until: *at(250)})
		require.NoError(t, err)
		require.Len(t, got, 2)

		byStep := map[string]metrics.WindowStep{}
		for _, ws := range got {
			byStep[ws.Step.Name] = ws
		}

		require.Equal(t, "CI", byStep["test"].Pipeline.Name)
		require.Equal(t, f.repo.ID, byStep["test"].Pipeline.RepoID)
		require.Equal(t, "failure", byStep["test"].Step.Conclusion)
		require.Equal(t, inside.ID, byStep["test"].Step.RunID)
		require.Equal(t, "Lint", byStep["golangci-lint"].Pipeline.Name)
	})

	t.Run("excludes steps of runs outside the window", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowSteps(context.Background(), metrics.RunWindowFilter{Since: *at(50), Until: *at(250)})
		require.NoError(t, err)

		for _, ws := range got {
			require.NotEqual(t, "build", ws.Step.Name)
		}
	})

	t.Run("filters by repo and forge", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowSteps(context.Background(), metrics.RunWindowFilter{RepoID: "nope", Since: *at(0), Until: *at(1000)})
		require.NoError(t, err)
		require.Empty(t, got)

		got, err = f.metrics.WindowSteps(context.Background(), metrics.RunWindowFilter{Forge: "forgejo", Since: *at(0), Until: *at(1000)})
		require.NoError(t, err)
		require.Empty(t, got)

		got, err = f.metrics.WindowSteps(context.Background(), metrics.RunWindowFilter{Forge: "github", Since: *at(0), Until: *at(1000)})
		require.NoError(t, err)
		require.Len(t, got, 3)
	})
}

func (f testFixture) seedRunWith(t *testing.T, run ingestion.Run) ingestion.Run {
	t.Helper()

	run.RepoID = f.repo.ID
	if run.PipelineName == "" {
		run.PipelineName = "CI"
	}

	if run.ForgeURL == "" {
		run.ForgeURL = "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + run.ForgeRunID
	}

	created, err := f.runs.UpsertRun(context.Background(), run)
	require.NoError(t, err)

	return created
}

func TestStore_ListRuns(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	oldest := f.seedRunWith(t, ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "success", StartedAt: at(10), CompletedAt: at(15)})
	failed := f.seedRunWith(t, ingestion.Run{
		ForgeRunID: "2", Status: "completed", Conclusion: "failure", StartedAt: at(20), CompletedAt: at(27),
		Branch: "main", SHA: "c4d291a", Message: "fix(stripe): webhook retry", Actor: "marcus-v",
	})
	timedOut := f.seedRunWith(t, ingestion.Run{ForgeRunID: "3", PipelineName: "Lint", Status: "completed", Conclusion: "timed_out", StartedAt: at(30), CompletedAt: at(40)})
	running := f.seedRunWith(t, ingestion.Run{ForgeRunID: "4", Status: "in_progress", StartedAt: at(50)})
	queued := f.seedRunWith(t, ingestion.Run{ForgeRunID: "5", Status: "queued"}) // no start time yet
	f.seedRunWith(t, ingestion.Run{ForgeRunID: "6", Status: "completed", Conclusion: "cancelled", StartedAt: at(5), CompletedAt: at(6)})

	// Two jobs in the failed run, the first started earlier: steps must come
	// back in job order, then step number.
	f.seedJobWithStep(t, failed.ID, "20", "checkout", 20, 21, 22, "success")
	f.seedJobWithStep(t, failed.ID, "21", "canary", 22, 23, 27, "failure")

	ids := func(runs []metrics.RunEntry) []string {
		out := make([]string, 0, len(runs))
		for _, r := range runs {
			out = append(out, r.ID)
		}

		return out
	}

	t.Run("lists newest first, with a run that hasn't started yet before the rest", func(t *testing.T) {
		t.Parallel()

		got, hasMore, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusAll})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Equal(t, []string{queued.ID, running.ID, timedOut.ID, failed.ID, oldest.ID}, ids(got)[:5])
		require.Len(t, got, 6)
	})

	t.Run("carries the run's pipeline, commit fields and forge link", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusFailed})
		require.NoError(t, err)

		var run metrics.RunEntry

		for _, r := range got {
			if r.ID == failed.ID {
				run = r
			}
		}

		require.Equal(t, "CI", run.Pipeline.Name)
		require.Equal(t, f.repo.ID, run.Pipeline.RepoID)
		require.Equal(t, "main", run.Branch)
		require.Equal(t, "c4d291a", run.SHA)
		require.Equal(t, "fix(stripe): webhook retry", run.Message)
		require.Equal(t, "marcus-v", run.Actor)
		require.Contains(t, run.ForgeURL, "/actions/runs/2")
	})

	t.Run("leaves commit fields empty when the run has none", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusSuccess})
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Empty(t, got[0].SHA)
		require.Empty(t, got[0].Actor)
	})

	t.Run("failed covers failure and timed_out, nothing else", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusFailed})
		require.NoError(t, err)
		require.Equal(t, []string{timedOut.ID, failed.ID}, ids(got))
	})

	t.Run("running covers every run that hasn't completed", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusRunning})
		require.NoError(t, err)
		require.Equal(t, []string{queued.ID, running.ID}, ids(got))
	})

	t.Run("success covers only successful conclusions", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusSuccess})
		require.NoError(t, err)
		require.Equal(t, []string{oldest.ID}, ids(got))
	})

	t.Run("attaches each run's steps in job order", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusFailed})
		require.NoError(t, err)

		var steps []metrics.RunStep

		for _, r := range got {
			if r.ID == failed.ID {
				steps = r.Steps
			}
		}

		require.Len(t, steps, 2)
		require.Equal(t, "checkout", steps[0].Name)
		require.Equal(t, "canary", steps[1].Name)
		require.Equal(t, "failure", steps[1].Conclusion)
	})

	t.Run("a run with no recorded steps has an empty, non-nil step list", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusSuccess})
		require.NoError(t, err)
		require.NotNil(t, got[0].Steps)
		require.Empty(t, got[0].Steps)
	})

	t.Run("paginates and reports whether more remain", func(t *testing.T) {
		t.Parallel()

		first, hasMore, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusAll, Limit: 2})
		require.NoError(t, err)
		require.True(t, hasMore)
		require.Equal(t, []string{queued.ID, running.ID}, ids(first))

		last, hasMore, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{Status: metrics.RunStatusAll, Limit: 4, Offset: 4})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, last, 2)
	})

	t.Run("filters by repo and forge", func(t *testing.T) {
		t.Parallel()

		got, _, err := f.metrics.ListRuns(ctx, metrics.RunListFilter{RepoID: "nope", Status: metrics.RunStatusAll})
		require.NoError(t, err)
		require.Empty(t, got)

		got, _, err = f.metrics.ListRuns(ctx, metrics.RunListFilter{Forge: "forgejo", Status: metrics.RunStatusAll})
		require.NoError(t, err)
		require.Empty(t, got)

		got, _, err = f.metrics.ListRuns(ctx, metrics.RunListFilter{Forge: "github", Status: metrics.RunStatusAll})
		require.NoError(t, err)
		require.Len(t, got, 6)
	})
}

package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/stretchr/testify/require"
)

func TestStore_RepoByIdentifier(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_token",
	})
	require.NoError(t, err)

	t.Run("finds the repo", func(t *testing.T) {
		t.Parallel()

		got, err := store.RepoByIdentifier(ctx, ingestion.ForgeGitHub, "alrayyes/pipeline-analytics")
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
	})

	t.Run("not found for a different forge", func(t *testing.T) {
		t.Parallel()

		_, err := store.RepoByIdentifier(ctx, ingestion.ForgeForgejo, "alrayyes/pipeline-analytics")
		require.ErrorIs(t, err, sqlite.ErrRepoNotFound)
	})
}

func TestStore_UpsertRun(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	repo, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "a/b", Token: "t"})
	require.NoError(t, err)

	started := time.Now().UTC().Truncate(time.Second)

	run := ingestion.Run{
		RepoID:       repo.ID,
		ForgeRunID:   "1001",
		PipelineName: "ci.yml",
		Status:       "in_progress",
		StartedAt:    &started,
		ForgeURL:     "https://github.com/a/b/actions/runs/1001",
	}

	t.Run("creates a new run", func(t *testing.T) {
		t.Parallel()

		created, err := store.UpsertRun(ctx, run)
		require.NoError(t, err)
		require.NotEmpty(t, created.ID)
		require.Equal(t, "in_progress", created.Status)
	})

	t.Run("updates the same run on a second delivery", func(t *testing.T) {
		t.Parallel()

		first, err := store.UpsertRun(ctx, ingestion.Run{
			RepoID: repo.ID, ForgeRunID: "1002", PipelineName: "ci.yml", Status: "in_progress",
		})
		require.NoError(t, err)

		completed := time.Now().UTC().Truncate(time.Second)
		second, err := store.UpsertRun(ctx, ingestion.Run{
			RepoID: repo.ID, ForgeRunID: "1002", PipelineName: "ci.yml",
			Status: "completed", Conclusion: "success", CompletedAt: &completed,
		})
		require.NoError(t, err)

		require.Equal(t, first.ID, second.ID)
		require.Equal(t, "completed", second.Status)
		require.Equal(t, "success", second.Conclusion)
	})
}

func TestStore_UpsertJobAndReplaceSteps(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	repo, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "a/b", Token: "t"})
	require.NoError(t, err)
	run, err := store.UpsertRun(ctx, ingestion.Run{RepoID: repo.ID, ForgeRunID: "1", PipelineName: "ci.yml", Status: "in_progress"})
	require.NoError(t, err)

	job := ingestion.Job{RunID: run.ID, ForgeJobID: "10", Name: "build", Status: "in_progress", ForgeURL: "https://github.com/a/b/actions/runs/1/job/10"}

	t.Run("creates then updates the same job", func(t *testing.T) {
		t.Parallel()

		first, err := store.UpsertJob(ctx, job)
		require.NoError(t, err)

		second, err := store.UpsertJob(ctx, ingestion.Job{RunID: run.ID, ForgeJobID: "10", Name: "build", Status: "completed", Conclusion: "failure", ForgeURL: job.ForgeURL})
		require.NoError(t, err)

		require.Equal(t, first.ID, second.ID)
		require.Equal(t, "completed", second.Status)
	})

	t.Run("replace steps overwrites the previous set", func(t *testing.T) {
		t.Parallel()

		j, err := store.UpsertJob(ctx, ingestion.Job{RunID: run.ID, ForgeJobID: "20", Name: "test", Status: "in_progress"})
		require.NoError(t, err)

		require.NoError(t, store.ReplaceSteps(ctx, j.ID, []ingestion.Step{
			{Number: 1, Name: "checkout", Status: "completed", Conclusion: "success"},
			{Number: 2, Name: "run tests", Status: "in_progress"},
		}))

		require.NoError(t, store.ReplaceSteps(ctx, j.ID, []ingestion.Step{
			{Number: 1, Name: "checkout", Status: "completed", Conclusion: "success"},
			{Number: 2, Name: "run tests", Status: "completed", Conclusion: "failure"},
		}))

		var count int
		err = store.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM steps WHERE job_id = ?", j.ID).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)

		var conclusion string
		err = store.DB().QueryRowContext(ctx, "SELECT conclusion FROM steps WHERE job_id = ? AND number = 2", j.ID).Scan(&conclusion)
		require.NoError(t, err)
		require.Equal(t, "failure", conclusion)
	})
}

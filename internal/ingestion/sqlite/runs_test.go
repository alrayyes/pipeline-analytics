package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/db"
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

func TestStore_UpsertRun_CommitMetadata(t *testing.T) {
	t.Parallel()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	store := sqlite.NewStore(conn, make([]byte, 32))
	ctx := context.Background()
	repo, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "a/b", Token: "t"})
	require.NoError(t, err)

	type commitRow struct{ branch, sha, message, actor sql.NullString }

	read := func(forgeRunID string) commitRow {
		var row commitRow

		require.NoError(t, conn.QueryRowContext(ctx,
			"SELECT branch, head_sha, head_message, actor FROM runs WHERE repo_id = ? AND forge_run_id = ?",
			repo.ID, forgeRunID,
		).Scan(&row.branch, &row.sha, &row.message, &row.actor))

		return row
	}

	t.Run("persists the commit fields", func(t *testing.T) {
		t.Parallel()

		_, err := store.UpsertRun(ctx, ingestion.Run{
			RepoID: repo.ID, ForgeRunID: "2001", PipelineName: "ci.yml", Status: "completed",
			Branch: "main", SHA: "c4d291a", Message: "fix: retry", Actor: "marcus-v",
		})
		require.NoError(t, err)

		row := read("2001")
		require.Equal(t, "main", row.branch.String)
		require.Equal(t, "c4d291a", row.sha.String)
		require.Equal(t, "fix: retry", row.message.String)
		require.Equal(t, "marcus-v", row.actor.String)
	})

	t.Run("stores null, not empty strings, when the forge omits them", func(t *testing.T) {
		t.Parallel()

		_, err := store.UpsertRun(ctx, ingestion.Run{RepoID: repo.ID, ForgeRunID: "2002", PipelineName: "ci.yml", Status: "completed"})
		require.NoError(t, err)

		row := read("2002")
		require.False(t, row.branch.Valid)
		require.False(t, row.actor.Valid)
	})

	t.Run("a later delivery without commit fields keeps the stored ones", func(t *testing.T) {
		t.Parallel()

		_, err := store.UpsertRun(ctx, ingestion.Run{
			RepoID: repo.ID, ForgeRunID: "2003", PipelineName: "ci.yml", Status: "in_progress",
			Branch: "main", SHA: "e1803bf", Message: "feat: pkce", Actor: "alex-r",
		})
		require.NoError(t, err)

		// A workflow_job event's stub run carries no commit data.
		_, err = store.UpsertRun(ctx, ingestion.Run{RepoID: repo.ID, ForgeRunID: "2003", PipelineName: "ci.yml", Status: "in_progress"})
		require.NoError(t, err)

		row := read("2003")
		require.Equal(t, "e1803bf", row.sha.String)
		require.Equal(t, "alex-r", row.actor.String)
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

func TestStore_RunStates(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	repoA, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "a/b", Token: "t"})
	require.NoError(t, err)
	repoB, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "c/d", Token: "t"})
	require.NoError(t, err)

	_, err = store.UpsertRun(ctx, ingestion.Run{
		RepoID: repoA.ID, ForgeRunID: "1001", PipelineName: "ci.yml",
		Status: "completed", Conclusion: "success",
	})
	require.NoError(t, err)
	_, err = store.UpsertRun(ctx, ingestion.Run{
		RepoID: repoA.ID, ForgeRunID: "1002", PipelineName: "ci.yml",
		Status: "in_progress",
	})
	require.NoError(t, err)
	_, err = store.UpsertRun(ctx, ingestion.Run{
		RepoID: repoB.ID, ForgeRunID: "2001", PipelineName: "ci.yml",
		Status: "completed", Conclusion: "failure",
	})
	require.NoError(t, err)

	t.Run("returns every stored run for the repo, keyed by forge run id", func(t *testing.T) {
		t.Parallel()

		states, err := store.RunStates(ctx, repoA.ID)
		require.NoError(t, err)
		require.Equal(t, map[string]ingestion.RunState{
			"1001": {Status: "completed", Conclusion: "success"},
			"1002": {Status: "in_progress"},
		}, states)
	})

	t.Run("doesn't include another repo's runs", func(t *testing.T) {
		t.Parallel()

		states, err := store.RunStates(ctx, repoA.ID)
		require.NoError(t, err)
		require.NotContains(t, states, "2001")
	})

	t.Run("a repo with no runs returns an empty map", func(t *testing.T) {
		t.Parallel()

		repoC, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "e/f", Token: "t"})
		require.NoError(t, err)

		states, err := store.RunStates(ctx, repoC.ID)
		require.NoError(t, err)
		require.Empty(t, states)
	})
}

func TestStore_RunByID(t *testing.T) {
	t.Parallel()

	t.Run("returns what an action needs", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)
		repo, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "a/b", Token: "t"})
		require.NoError(t, err)

		saved, err := store.UpsertRun(context.Background(), ingestion.Run{
			RepoID: repo.ID, ForgeRunID: "9", PipelineName: "CI", Status: "completed", Conclusion: "failure",
		})
		require.NoError(t, err)

		got, err := store.RunByID(context.Background(), saved.ID)
		require.NoError(t, err)
		require.Equal(t, "failure", got.Conclusion)
		require.Equal(t, "9", got.ForgeRunID)
	})

	t.Run("an unknown id is ErrRunNotFound", func(t *testing.T) {
		t.Parallel()

		_, err := newTestStore(t).RunByID(context.Background(), "missing")

		require.ErrorIs(t, err, ingestion.ErrRunNotFound)
	})
}

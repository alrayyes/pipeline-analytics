package sqlite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *sqlite.Store {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	key := make([]byte, 32)

	return sqlite.NewStore(conn, key)
}

func TestStore_CreateRepo(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	repo, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	t.Run("returns a masked token, never the raw value", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "****1234", repo.TokenMasked)
	})

	t.Run("stores the token encrypted, not in plaintext", func(t *testing.T) {
		t.Parallel()

		var encrypted []byte

		err := store.DB().QueryRowContext(ctx, "SELECT token_encrypted FROM repos WHERE id = ?", repo.ID).Scan(&encrypted)
		require.NoError(t, err)
		require.NotContains(t, string(encrypted), "ghp_supersecrettoken1234")
	})

	t.Run("generates a per-repo webhook secret", func(t *testing.T) {
		t.Parallel()

		require.NotEmpty(t, repo.WebhookSecret)
	})

	t.Run("starts pending", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, ingestion.StatusPending, repo.IngestionStatus)
	})
}

func TestStore_RepoToken(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	repo, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	token, err := store.RepoToken(ctx, repo.ID)
	require.NoError(t, err)
	require.Equal(t, "ghp_supersecrettoken1234", token)
}

func TestStore_ListAndGetRepo(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateRepo(ctx, ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
		Forge:              ingestion.ForgeForgejo,
		Identifier:         "alrayyes/dotfiles",
		ForgejoInstanceURL: "https://git.higherlearning.eu",
		Token:              "forgejo-token-5678",
	})
	require.NoError(t, err)

	t.Run("list includes the created repo", func(t *testing.T) {
		t.Parallel()

		repos, hasMore, err := store.ListRepos(ctx, ingestion.RepoListFilter{})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, repos, 1)
		require.Equal(t, created.ID, repos[0].ID)
		require.Equal(t, "****5678", repos[0].TokenMasked)
	})

	t.Run("get returns the same repo", func(t *testing.T) {
		t.Parallel()

		got, err := store.GetRepo(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.Identifier, got.Identifier)
	})
}

func TestStore_ListReposPagination(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	var created []ingestion.Repo

	for i := range 3 {
		repo, err := store.CreateRepo(ctx, ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: fmt.Sprintf("alrayyes/repo-%d", i),
			Token:      "ghp_supersecrettoken1234",
		})
		require.NoError(t, err)
		created = append(created, repo)
	}

	t.Run("limit trims the page and reports more remain", func(t *testing.T) {
		t.Parallel()

		repos, hasMore, err := store.ListRepos(ctx, ingestion.RepoListFilter{Limit: 2})
		require.NoError(t, err)
		require.True(t, hasMore)
		require.Len(t, repos, 2)
		require.Equal(t, created[0].ID, repos[0].ID)
		require.Equal(t, created[1].ID, repos[1].ID)
	})

	t.Run("offset returns the next page", func(t *testing.T) {
		t.Parallel()

		repos, hasMore, err := store.ListRepos(ctx, ingestion.RepoListFilter{Limit: 2, Offset: 2})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, repos, 1)
		require.Equal(t, created[2].ID, repos[0].ID)
	})

	t.Run("no limit returns every repo unpaginated", func(t *testing.T) {
		t.Parallel()

		repos, hasMore, err := store.ListRepos(ctx, ingestion.RepoListFilter{})
		require.NoError(t, err)
		require.False(t, hasMore)
		require.Len(t, repos, 3)
	})
}

func TestStore_ListReposForgeFilter(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	forgejoRepo, err := store.CreateRepo(ctx, ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
		Forge:              ingestion.ForgeForgejo,
		Identifier:         "alrayyes/dotfiles",
		ForgejoInstanceURL: "https://git.higherlearning.eu",
		Token:              "forgejo-token-5678",
	})
	require.NoError(t, err)

	repos, hasMore, err := store.ListRepos(ctx, ingestion.RepoListFilter{Forge: ingestion.ForgeForgejo})
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Len(t, repos, 1)
	require.Equal(t, forgejoRepo.ID, repos[0].ID)
}

func TestStore_SetIngestionStatus(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	require.NoError(t, store.SetIngestionStatus(ctx, created.ID, ingestion.StatusDegraded, "insufficient token scope"))

	got, err := store.GetRepo(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, ingestion.StatusDegraded, got.IngestionStatus)
	require.Equal(t, "insufficient token scope", got.IngestionStatusReason)
}

func TestStore_SetReconcileETag(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)
	require.Empty(t, created.ReconcileETag)

	require.NoError(t, store.SetReconcileETag(ctx, created.ID, `"v1"`))

	got, err := store.GetRepo(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, `"v1"`, got.ReconcileETag)
}

func TestStore_SetReconcileETag_UnknownRepo(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	err := store.SetReconcileETag(context.Background(), "does-not-exist", `"v1"`)
	require.ErrorIs(t, err, sqlite.ErrRepoNotFound)
}

func TestStore_DeleteRepo(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	require.NoError(t, store.DeleteRepo(ctx, created.ID))

	_, err = store.GetRepo(ctx, created.ID)
	require.ErrorIs(t, err, sqlite.ErrRepoNotFound)
}

func TestStore_DeleteRepo_UnknownRepo(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	err := store.DeleteRepo(context.Background(), "does-not-exist")
	require.ErrorIs(t, err, sqlite.ErrRepoNotFound)
}

func TestStore_CreateRepo_AlreadyTracked(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.CreateRepo(ctx, ingestion.NewRepo{
		Forge:      ingestion.ForgeGitHub,
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	t.Run("same forge and identifier, both GitHub (no instance URL)", func(t *testing.T) {
		t.Parallel()

		_, err := store.CreateRepo(ctx, ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_supersecrettoken1234",
		})
		require.ErrorIs(t, err, ingestion.ErrRepoAlreadyTracked)
	})

	t.Run("same identifier on a different forge is not a duplicate", func(t *testing.T) {
		t.Parallel()

		_, err := store.CreateRepo(ctx, ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
			Forge:              ingestion.ForgeForgejo,
			Identifier:         "alrayyes/pipeline-analytics",
			ForgejoInstanceURL: "https://git.higherlearning.eu",
			Token:              "forgejo-token-5678",
		})
		require.NoError(t, err)
	})
}

func TestStore_CreateRepo_AlreadyTracked_Forgejo(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.CreateRepo(ctx, ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
		Forge:              ingestion.ForgeForgejo,
		Identifier:         "alrayyes/dotfiles",
		ForgejoInstanceURL: "https://git.higherlearning.eu",
		Token:              "forgejo-token-5678",
	})
	require.NoError(t, err)

	t.Run("same forge, identifier, and instance URL", func(t *testing.T) {
		t.Parallel()

		_, err := store.CreateRepo(ctx, ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
			Forge:              ingestion.ForgeForgejo,
			Identifier:         "alrayyes/dotfiles",
			ForgejoInstanceURL: "https://git.higherlearning.eu",
			Token:              "forgejo-token-5678",
		})
		require.ErrorIs(t, err, ingestion.ErrRepoAlreadyTracked)
	})

	t.Run("same identifier on a different instance URL is not a duplicate", func(t *testing.T) {
		t.Parallel()

		_, err := store.CreateRepo(ctx, ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
			Forge:              ingestion.ForgeForgejo,
			Identifier:         "alrayyes/dotfiles",
			ForgejoInstanceURL: "https://code.example.com",
			Token:              "forgejo-token-9999",
		})
		require.NoError(t, err)
	})
}

func TestStore_ListRepoIdentifiers(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	// More than PAGE_SIZE (20) so this proves the result isn't paginated.
	for i := range 25 {
		_, err := store.CreateRepo(ctx, ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: fmt.Sprintf("alrayyes/repo-%d", i),
			Token:      "ghp_supersecrettoken1234",
		})
		require.NoError(t, err)
	}

	_, err := store.CreateRepo(ctx, ingestion.NewRepo{ //nolint:gosec // test fixture value, not a real credential
		Forge:              ingestion.ForgeForgejo,
		Identifier:         "alrayyes/dotfiles",
		ForgejoInstanceURL: "https://git.higherlearning.eu",
		Token:              "forgejo-token-5678",
	})
	require.NoError(t, err)

	t.Run("returns every identifier for the forge, unpaginated", func(t *testing.T) {
		t.Parallel()

		identifiers, err := store.ListRepoIdentifiers(ctx, ingestion.ForgeGitHub, "")
		require.NoError(t, err)
		require.Len(t, identifiers, 25)
		require.Contains(t, identifiers, "alrayyes/repo-24")
	})

	t.Run("scopes to the forge and instance URL", func(t *testing.T) {
		t.Parallel()

		identifiers, err := store.ListRepoIdentifiers(ctx, ingestion.ForgeForgejo, "https://git.higherlearning.eu")
		require.NoError(t, err)
		require.Equal(t, []string{"alrayyes/dotfiles"}, identifiers)
	})

	t.Run("a different instance URL sees nothing", func(t *testing.T) {
		t.Parallel()

		identifiers, err := store.ListRepoIdentifiers(ctx, ingestion.ForgeForgejo, "https://code.example.com")
		require.NoError(t, err)
		require.Empty(t, identifiers)
	})
}

func TestStore_LocateJob(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	repo, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "o/n", Token: "t"})
	require.NoError(t, err)

	run, err := store.UpsertRun(ctx, ingestion.Run{RepoID: repo.ID, ForgeRunID: "1", PipelineName: "CI", Status: "completed"})
	require.NoError(t, err)

	job, err := store.UpsertJob(ctx, ingestion.Job{RunID: run.ID, ForgeJobID: "5001", Name: "build", Status: "completed", ForgeURL: "https://forge/job/5001"})
	require.NoError(t, err)

	t.Run("finds a job by its run and id", func(t *testing.T) {
		t.Parallel()

		got, err := store.LocateJob(ctx, run.ID, job.ID)
		require.NoError(t, err)
		require.Equal(t, ingestion.JobLocation{RepoID: repo.ID, ForgeJobID: "5001", ForgeURL: "https://forge/job/5001"}, got)
	})

	t.Run("a job under another run is not found", func(t *testing.T) {
		t.Parallel()

		_, err := store.LocateJob(ctx, "other", job.ID)
		require.ErrorIs(t, err, ingestion.ErrJobNotFound)
	})

	t.Run("an unknown job is not found", func(t *testing.T) {
		t.Parallel()

		_, err := store.LocateJob(ctx, run.ID, "nope")
		require.ErrorIs(t, err, ingestion.ErrJobNotFound)
	})
}

func TestStore_ForgeTokens(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("saves a token and reads it back for the same forge", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)

		saved, err := store.SaveForgeToken(ctx, ingestion.ForgeGitHub, "", "ghp_secret1234")
		require.NoError(t, err)
		require.NotEmpty(t, saved.ID)
		require.Equal(t, "****1234", saved.TokenMasked)

		got, err := store.ForgeToken(ctx, ingestion.ForgeGitHub, "")
		require.NoError(t, err)
		require.Equal(t, "ghp_secret1234", got)
	})

	t.Run("never keeps the token in the clear", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)

		_, err := store.SaveForgeToken(ctx, ingestion.ForgeGitHub, "", "ghp_secret1234")
		require.NoError(t, err)

		var stored []byte
		require.NoError(t, store.DB().QueryRowContext(ctx, "SELECT token_encrypted FROM forge_tokens").Scan(&stored))
		require.NotContains(t, string(stored), "ghp_secret1234")
	})

	t.Run("saving again for the same forge replaces it, keeping one row", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)

		_, err := store.SaveForgeToken(ctx, ingestion.ForgeGitHub, "", "ghp_old00000")
		require.NoError(t, err)

		saved, err := store.SaveForgeToken(ctx, ingestion.ForgeGitHub, "", "ghp_new11111")
		require.NoError(t, err)
		require.Equal(t, "****1111", saved.TokenMasked)

		got, err := store.ForgeToken(ctx, ingestion.ForgeGitHub, "")
		require.NoError(t, err)
		require.Equal(t, "ghp_new11111", got)

		list, err := store.ListForgeTokens(ctx)
		require.NoError(t, err)
		require.Len(t, list, 1)
	})

	t.Run("Forgejo tokens are per instance, and a trailing slash is the same instance", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)

		_, err := store.SaveForgeToken(ctx, ingestion.ForgeForgejo, "https://git.example/", "fj_aaaa1111")
		require.NoError(t, err)

		_, err = store.SaveForgeToken(ctx, ingestion.ForgeForgejo, "https://other.example", "fj_bbbb2222")
		require.NoError(t, err)

		got, err := store.ForgeToken(ctx, ingestion.ForgeForgejo, "https://git.example")
		require.NoError(t, err)
		require.Equal(t, "fj_aaaa1111", got)

		list, err := store.ListForgeTokens(ctx)
		require.NoError(t, err)
		require.Len(t, list, 2)
	})

	t.Run("an instance URL means nothing for GitHub", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)

		_, err := store.SaveForgeToken(ctx, ingestion.ForgeGitHub, "https://ignored.example", "ghp_secret1234")
		require.NoError(t, err)

		got, err := store.ForgeToken(ctx, ingestion.ForgeGitHub, "")
		require.NoError(t, err)
		require.Equal(t, "ghp_secret1234", got)
	})

	t.Run("no token saved is not found", func(t *testing.T) {
		t.Parallel()

		_, err := newTestStore(t).ForgeToken(ctx, ingestion.ForgeGitHub, "")
		require.ErrorIs(t, err, ingestion.ErrSavedTokenNotFound)
	})

	t.Run("deleting removes it, and an unknown id is not found", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)

		saved, err := store.SaveForgeToken(ctx, ingestion.ForgeGitHub, "", "ghp_secret1234")
		require.NoError(t, err)

		require.NoError(t, store.DeleteForgeToken(ctx, saved.ID))

		_, err = store.ForgeToken(ctx, ingestion.ForgeGitHub, "")
		require.ErrorIs(t, err, ingestion.ErrSavedTokenNotFound)

		require.ErrorIs(t, store.DeleteForgeToken(ctx, saved.ID), ingestion.ErrSavedTokenNotFound)
	})

	t.Run("deleting a saved token leaves a registered repo's own copy working", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)

		saved, err := store.SaveForgeToken(ctx, ingestion.ForgeGitHub, "", "ghp_secret1234")
		require.NoError(t, err)

		repo, err := store.CreateRepo(ctx, ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "o/n", Token: "ghp_secret1234"})
		require.NoError(t, err)

		require.NoError(t, store.DeleteForgeToken(ctx, saved.ID))

		got, err := store.RepoToken(ctx, repo.ID)
		require.NoError(t, err)
		require.Equal(t, "ghp_secret1234", got)
	})
}

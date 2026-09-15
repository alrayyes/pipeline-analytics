package sqlite_test

import (
	"context"
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

		repos, err := store.ListRepos(ctx)
		require.NoError(t, err)
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

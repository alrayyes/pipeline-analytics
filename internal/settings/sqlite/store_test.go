package sqlite_test

import (
	"context"
	"testing"

	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/settings/sqlite"
	"github.com/stretchr/testify/require"
)

// newTestStore returns a settings Store plus a real user ID for it to
// reference -- account_settings.user_id REFERENCES webauthn_users(id).
func newTestStore(t *testing.T) (*sqlite.Store, string) {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	user, err := authsqlite.NewStore(conn).CreateUser(context.Background(), []byte("handle-1"), "admin")
	require.NoError(t, err)

	return sqlite.NewStore(conn), user.ID
}

func TestStore_Get(t *testing.T) {
	t.Parallel()

	t.Run("a user with no settings row yet returns an empty map, not an error", func(t *testing.T) {
		t.Parallel()

		store, userID := newTestStore(t)

		got, err := store.Get(context.Background(), userID)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("returns whatever was previously patched", func(t *testing.T) {
		t.Parallel()

		store, userID := newTestStore(t)
		ctx := context.Background()

		_, err := store.Patch(ctx, userID, map[string]*string{"theme": new("dark")})
		require.NoError(t, err)

		got, err := store.Get(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"theme": "dark"}, got)
	})
}

func TestStore_Patch(t *testing.T) {
	t.Parallel()

	t.Run("sets a new key", func(t *testing.T) {
		t.Parallel()

		store, userID := newTestStore(t)

		got, err := store.Patch(context.Background(), userID, map[string]*string{"theme": new("dark")})
		require.NoError(t, err)
		require.Equal(t, map[string]string{"theme": "dark"}, got)
	})

	t.Run("updates an existing key without disturbing others", func(t *testing.T) {
		t.Parallel()

		store, userID := newTestStore(t)
		ctx := context.Background()

		_, err := store.Patch(ctx, userID, map[string]*string{
			"theme":       new("dark"),
			"forgeFilter": new("github"),
		})
		require.NoError(t, err)

		got, err := store.Patch(ctx, userID, map[string]*string{"theme": new("light")})
		require.NoError(t, err)
		require.Equal(t, map[string]string{"theme": "light", "forgeFilter": "github"}, got)
	})

	t.Run("a nil value clears the key entirely", func(t *testing.T) {
		t.Parallel()

		store, userID := newTestStore(t)
		ctx := context.Background()

		_, err := store.Patch(ctx, userID, map[string]*string{
			"theme":       new("dark"),
			"forgeFilter": new("github"),
		})
		require.NoError(t, err)

		got, err := store.Patch(ctx, userID, map[string]*string{"theme": nil})
		require.NoError(t, err)
		require.Equal(t, map[string]string{"forgeFilter": "github"}, got)
	})

	t.Run("persists across a fresh Get, not just the Patch return value", func(t *testing.T) {
		t.Parallel()

		store, userID := newTestStore(t)
		ctx := context.Background()

		_, err := store.Patch(ctx, userID, map[string]*string{"theme": new("dark")})
		require.NoError(t, err)

		got, err := store.Get(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"theme": "dark"}, got)
	})
}

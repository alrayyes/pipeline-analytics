package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	t.Parallel()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	require.NoError(t, db.Migrate(context.Background(), conn))

	tables := []string{"repos", "runs", "jobs", "steps"}
	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			t.Parallel()

			var name string
			err := conn.QueryRow(
				"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table,
			).Scan(&name)
			require.NoError(t, err)
			require.Equal(t, table, name)
		})
	}
}

// TestMigrate_DedupesPreExistingRepos reproduces the pre-#199 state --
// migrations 1-4 applied, a duplicate GitHub repo (both rows with a NULL
// forgejo_instance_url) inserted before migration 5's unique index
// existed to stop it -- and pins down that migration 5 now cleans up such
// duplicates instead of failing to apply (#247).
func TestMigrate_DedupesPreExistingRepos(t *testing.T) {
	t.Parallel()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	ctx := context.Background()

	provider, err := goose.NewProvider(goose.DialectSQLite3, conn, os.DirFS("migrations"))
	require.NoError(t, err)
	_, err = provider.UpTo(ctx, 4)
	require.NoError(t, err)

	insertRepo := func(id, createdAt string) {
		t.Helper()

		_, err := conn.ExecContext(ctx, `
			INSERT INTO repos (id, forge, identifier, token_encrypted, token_masked, webhook_secret, ingestion_status, created_at)
			VALUES (?, 'github', 'alrayyes/pipeline-analytics', x'00', '****1234', 'secret', 'active', ?)
		`, id, createdAt)
		require.NoError(t, err)
	}

	insertRepo("older", "2026-01-01 00:00:00")
	insertRepo("newer", "2026-01-02 00:00:00")

	require.NoError(t, db.Migrate(ctx, conn))

	var count int
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM repos").Scan(&count))
	require.Equal(t, 1, count)

	var remainingID string
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT id FROM repos").Scan(&remainingID))
	require.Equal(t, "older", remainingID)
}

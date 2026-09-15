package db_test

import (
	"context"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/db"
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

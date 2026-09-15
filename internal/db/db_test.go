package db_test

import (
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/stretchr/testify/require"
)

func TestOpenInMemory(t *testing.T) {
	t.Parallel()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	require.NoError(t, conn.Ping())
}

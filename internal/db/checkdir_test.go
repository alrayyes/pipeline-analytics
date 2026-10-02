package db_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/stretchr/testify/require"
)

func TestCheckDir(t *testing.T) {
	t.Parallel()

	t.Run("passes for a writable directory", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, db.CheckDir(filepath.Join(t.TempDir(), "app.db")))
	})

	t.Run("leaves nothing behind in the directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, db.CheckDir(filepath.Join(dir, "app.db")))

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("an in-memory database has no directory to check", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, db.CheckDir(":memory:"))
	})

	t.Run("a bare file name checks the working directory", func(t *testing.T) {
		t.Parallel()

		// The package's own directory is writable under go test.
		require.NoError(t, db.CheckDir("app.db"))
	})

	t.Run("names the directory when it does not exist", func(t *testing.T) {
		t.Parallel()

		dir := filepath.Join(t.TempDir(), "missing")

		err := db.CheckDir(filepath.Join(dir, "app.db"))

		require.ErrorIs(t, err, db.ErrDirMissing)
		require.ErrorContains(t, err, dir)
	})

	t.Run("names the directory when the path's parent is a file", func(t *testing.T) {
		t.Parallel()

		file := filepath.Join(t.TempDir(), "not-a-dir")
		require.NoError(t, os.WriteFile(file, nil, 0o600))

		err := db.CheckDir(filepath.Join(file, "app.db"))

		require.ErrorIs(t, err, db.ErrNotDirectory)
		require.ErrorContains(t, err, file)
	})

	t.Run("says the directory is not writable, and by whom, naming the directory", func(t *testing.T) {
		t.Parallel()

		if os.Geteuid() == 0 {
			t.Skip("root can write anywhere, so a read-only directory proves nothing")
		}

		dir := t.TempDir()
		// An owner-read-only directory is the whole point of this test.
		require.NoError(t, os.Chmod(dir, 0o500))       //nolint:gosec // G302: see above
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) //nolint:gosec // G302: restores the temp dir so it can be removed

		err := db.CheckDir(filepath.Join(dir, "app.db"))

		require.ErrorContains(t, err, dir)
		require.ErrorContains(t, err, "isn't writable")
		require.ErrorContains(t, err, "uid")
	})
}

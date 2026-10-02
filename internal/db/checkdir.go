package db

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// ErrDirMissing and ErrNotDirectory are what CheckDir wraps for a database
// path whose directory can't be used, so callers and tests match on the kind
// of fault with errors.Is rather than on message text.
var (
	ErrDirMissing   = errors.New("does not exist")
	ErrNotDirectory = errors.New("is not a directory")
)

// CheckDir reports whether a SQLite database at path can be created or
// opened: its directory has to exist and be writable by this process, since
// SQLite writes a journal beside the file. Without it, a data directory owned
// by another user fails at startup as "unable to open database file (14)",
// which names neither the directory nor the cause. An in-memory database has
// no directory to check.
func CheckDir(path string) error {
	if path == ":memory:" || path == "" {
		return nil
	}

	dir := filepath.Dir(path)

	info, err := os.Stat(dir)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("data directory %s %w", dir, ErrDirMissing)
	case err != nil:
		return fmt.Errorf("check data directory %s: %w", dir, err)
	case !info.IsDir():
		return fmt.Errorf("data directory %s %w", dir, ErrNotDirectory)
	}

	probe, err := os.CreateTemp(dir, ".write-check-*")
	if err != nil {
		return fmt.Errorf("data directory %s isn't writable by uid %d (it must be owned by that user, or writable by it): %w",
			dir, os.Geteuid(), err)
	}

	name := probe.Name()

	if err := probe.Close(); err != nil {
		return fmt.Errorf("close write check in %s: %w", dir, err)
	}

	if err := os.Remove(name); err != nil && !errors.Is(err, syscall.ENOENT) {
		return fmt.Errorf("remove write check in %s: %w", dir, err)
	}

	return nil
}

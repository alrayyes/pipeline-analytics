// Package db opens the pipeline-analytics SQLite database.
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// Open opens the SQLite database at path (or ":memory:") with the pragmas
// pipeline-analytics needs: foreign keys enforced, and a busy timeout so
// concurrent writers back off instead of failing immediately.
func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	return conn, nil
}

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
//
// The pool is capped at a single connection: SQLite serializes writers
// regardless, and a ":memory:" DSN backs a separate, empty database per
// connection under database/sql's pool -- a second pooled connection would
// silently see none of the first connection's tables or data.
func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	conn.SetMaxOpenConns(1)

	return conn, nil
}

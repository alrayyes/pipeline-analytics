// Package sqlite is the SQLite-backed adapter for the settings.Store port.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// Store implements settings.Store against a SQLite database.
type Store struct {
	db *sql.DB
}

// NewStore returns a Store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Get implements settings.Store.
func (s *Store) Get(ctx context.Context, userID string) (map[string]string, error) {
	var data []byte

	err := s.db.QueryRowContext(ctx, "SELECT data FROM account_settings WHERE user_id = ?", userID).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return map[string]string{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("query account settings: %w", err)
	}

	return decode(data)
}

// Patch implements settings.Store.
func (s *Store) Patch(ctx context.Context, userID string, updates map[string]*string) (map[string]string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var existing []byte

	err = tx.QueryRowContext(ctx, "SELECT data FROM account_settings WHERE user_id = ?", userID).Scan(&existing)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("query account settings: %w", err)
	}

	overrides, err := decode(existing)
	if err != nil {
		return nil, err
	}

	for key, value := range updates {
		if value == nil {
			delete(overrides, key)
		} else {
			overrides[key] = *value
		}
	}

	encoded, err := json.Marshal(overrides)
	if err != nil {
		return nil, fmt.Errorf("encode account settings: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO account_settings (user_id, data, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at
	`, userID, encoded)
	if err != nil {
		return nil, fmt.Errorf("upsert account settings: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit account settings: %w", err)
	}

	return overrides, nil
}

// decode parses a stored settings blob, treating both SQL NULL (no row
// yet) and an empty column as no overrides at all, rather than a JSON
// error.
func decode(data []byte) (map[string]string, error) {
	if len(data) == 0 {
		return map[string]string{}, nil
	}

	var overrides map[string]string
	if err := json.Unmarshal(data, &overrides); err != nil {
		return nil, fmt.Errorf("decode account settings: %w", err)
	}

	if overrides == nil {
		overrides = map[string]string{}
	}

	return overrides, nil
}

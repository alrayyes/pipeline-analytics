// Package sqlite is the SQLite-backed adapter for the auth.Store port.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/crypto"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// sessionTTL is how long a session lasts after creation. A solo-dev tool
// with no password-reset flow to fall back on favors a long, low-friction
// lifetime over frequent re-authentication.
const sessionTTL = 30 * 24 * time.Hour

// Store implements auth.Store against a SQLite database.
type Store struct {
	db *sql.DB
}

// NewStore returns a Store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// CreateUser implements auth.Store.
func (s *Store) CreateUser(ctx context.Context, userHandle []byte, displayName string) (auth.User, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM webauthn_users").Scan(&count); err != nil {
		return auth.User{}, fmt.Errorf("count users: %w", err)
	}

	if count > 0 {
		return auth.User{}, auth.ErrUserExists
	}

	user := auth.User{
		ID:          uuid.NewString(),
		UserHandle:  userHandle,
		DisplayName: displayName,
		CreatedAt:   time.Now().UTC(),
	}

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO webauthn_users (id, user_handle, display_name, created_at) VALUES (?, ?, ?, ?)",
		user.ID, user.UserHandle, user.DisplayName, user.CreatedAt,
	)
	if err != nil {
		return auth.User{}, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

// GetUser implements auth.Store.
func (s *Store) GetUser(ctx context.Context) (auth.User, error) {
	var user auth.User

	err := s.db.QueryRowContext(ctx, "SELECT id, user_handle, display_name, created_at FROM webauthn_users LIMIT 1").
		Scan(&user.ID, &user.UserHandle, &user.DisplayName, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrNoUser
	}

	if err != nil {
		return auth.User{}, fmt.Errorf("query user: %w", err)
	}

	return user, nil
}

// PutCredential implements auth.Store.
func (s *Store) PutCredential(ctx context.Context, userID string, cred webauthn.Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("marshal credential: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO webauthn_credentials (id, user_id, data, created_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET data = excluded.data
	`, cred.ID, userID, data, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("upsert credential: %w", err)
	}

	return nil
}

// CredentialsForUser implements auth.Store.
func (s *Store) CredentialsForUser(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT data FROM webauthn_credentials WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("query credentials: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var creds []webauthn.Credential

	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan credential: %w", err)
		}

		var cred webauthn.Credential
		if err := json.Unmarshal(data, &cred); err != nil {
			return nil, fmt.Errorf("unmarshal credential: %w", err)
		}

		creds = append(creds, cred)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credentials: %w", err)
	}

	return creds, nil
}

// SaveCeremony implements auth.Store.
func (s *Store) SaveCeremony(ctx context.Context, id string, data []byte, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO webauthn_ceremonies (id, data, expires_at) VALUES (?, ?, ?)",
		id, data, expiresAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert ceremony: %w", err)
	}

	return nil
}

// TakeCeremony implements auth.Store.
func (s *Store) TakeCeremony(ctx context.Context, id string) ([]byte, error) {
	var (
		data      []byte
		expiresAt time.Time
	)

	err := s.db.QueryRowContext(ctx, "SELECT data, expires_at FROM webauthn_ceremonies WHERE id = ?", id).Scan(&data, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, auth.ErrCeremonyNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("query ceremony: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, "DELETE FROM webauthn_ceremonies WHERE id = ?", id); err != nil {
		return nil, fmt.Errorf("delete ceremony: %w", err)
	}

	if time.Now().After(expiresAt) {
		return nil, auth.ErrCeremonyNotFound
	}

	return data, nil
}

// CreateSession implements auth.Store.
func (s *Store) CreateSession(ctx context.Context, userID string) (string, error) {
	id, err := crypto.RandomHex(32)
	if err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}

	now := time.Now().UTC()

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES (?, ?, ?, ?)",
		id, userID, now.Add(sessionTTL), now,
	)
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}

	return id, nil
}

// Session implements auth.Store.
func (s *Store) Session(ctx context.Context, sessionID string) (string, error) {
	var (
		userID    string
		expiresAt time.Time
	)

	err := s.db.QueryRowContext(ctx, "SELECT user_id, expires_at FROM sessions WHERE id = ?", sessionID).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", auth.ErrSessionNotFound
	}

	if err != nil {
		return "", fmt.Errorf("query session: %w", err)
	}

	if time.Now().After(expiresAt) {
		return "", auth.ErrSessionNotFound
	}

	return userID, nil
}

// DeleteSession implements auth.Store.
func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

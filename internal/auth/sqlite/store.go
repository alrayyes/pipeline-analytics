// Package sqlite is the SQLite-backed adapter for the auth.Store port.
package sqlite

import (
	"context"
	"crypto/sha256"
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

// apiTokenTTL is how long an API token lasts after creation, per
// add-api-token-auth/design.md's "a long, fixed TTL rather than no expiry
// at all" decision: long enough a script won't need to babysit rotation,
// short enough a forgotten token doesn't stay valid forever. Revocation is
// the mechanism for anything sooner; this is the backstop.
const apiTokenTTL = 365 * 24 * time.Hour

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

// PutCredentialLabeled implements auth.Store.
func (s *Store) PutCredentialLabeled(ctx context.Context, userID string, cred webauthn.Credential, label string) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("marshal credential: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO webauthn_credentials (id, user_id, data, label, created_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET data = excluded.data
	`, cred.ID, userID, data, label, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert labeled credential: %w", err)
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

// CredentialInfosForUser implements auth.Store.
func (s *Store) CredentialInfosForUser(ctx context.Context, userID string) ([]auth.CredentialInfo, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, label, created_at FROM webauthn_credentials WHERE user_id = ? ORDER BY created_at",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query credential infos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var infos []auth.CredentialInfo

	for rows.Next() {
		var info auth.CredentialInfo
		if err := rows.Scan(&info.ID, &info.Label, &info.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan credential info: %w", err)
		}

		infos = append(infos, info)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credential infos: %w", err)
	}

	return infos, nil
}

// RevokeCredential implements auth.Store.
func (s *Store) RevokeCredential(ctx context.Context, userID string, credentialID []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM webauthn_credentials WHERE user_id = ?", userID).Scan(&count); err != nil {
		return fmt.Errorf("count credentials: %w", err)
	}

	if count <= 1 {
		return auth.ErrLastCredential
	}

	res, err := tx.ExecContext(ctx,
		"DELETE FROM webauthn_credentials WHERE id = ? AND user_id = ?", credentialID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete credential: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete credential: %w", err)
	}

	if n == 0 {
		return auth.ErrCredentialNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
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

// CreateToken implements auth.Store.
func (s *Store) CreateToken(ctx context.Context, userID string) (auth.Token, string, error) {
	raw, err := crypto.RandomHex(32)
	if err != nil {
		return auth.Token{}, "", fmt.Errorf("generate token: %w", err)
	}

	hash := tokenHash(raw)
	now := time.Now().UTC()

	tok := auth.Token{
		ID:        uuid.NewString(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(apiTokenTTL),
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO api_tokens (id, user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?)",
		tok.ID, tok.UserID, hash, tok.ExpiresAt, tok.CreatedAt,
	)
	if err != nil {
		return auth.Token{}, "", fmt.Errorf("insert token: %w", err)
	}

	return tok, raw, nil
}

// TokenUserID implements auth.Store.
func (s *Store) TokenUserID(ctx context.Context, rawToken string) (string, error) {
	var (
		userID    string
		expiresAt time.Time
		revokedAt sql.NullTime
	)

	err := s.db.QueryRowContext(ctx,
		"SELECT user_id, expires_at, revoked_at FROM api_tokens WHERE token_hash = ?",
		tokenHash(rawToken),
	).Scan(&userID, &expiresAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", auth.ErrTokenNotFound
	}

	if err != nil {
		return "", fmt.Errorf("query token: %w", err)
	}

	if revokedAt.Valid || time.Now().After(expiresAt) {
		return "", auth.ErrTokenNotFound
	}

	return userID, nil
}

// RevokeToken implements auth.Store.
func (s *Store) RevokeToken(ctx context.Context, userID, tokenID string) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE api_tokens SET revoked_at = ? WHERE id = ? AND user_id = ? AND revoked_at IS NULL",
		time.Now().UTC(), tokenID, userID,
	)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	if n == 0 {
		return auth.ErrTokenNotFound
	}

	return nil
}

// tokenHash hashes a raw API token for storage/lookup, so the raw secret
// itself is never recoverable from the database. See
// add-api-token-auth/design.md's "Hash the token at rest" decision.
func tokenHash(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))

	return sum[:]
}

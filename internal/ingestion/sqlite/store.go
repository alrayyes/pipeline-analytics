// Package sqlite is the SQLite-backed adapter for the ingestion.Store port.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/crypto"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/google/uuid"
	sqlitedriver "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// Store implements ingestion.Store against a SQLite database.
type Store struct {
	db  *sql.DB
	key []byte
}

// NewStore returns a Store that encrypts tokens at rest under key (32 bytes,
// AES-256).
func NewStore(db *sql.DB, key []byte) *Store {
	return &Store{db: db, key: key}
}

// DB returns the underlying connection, for callers (such as tests) that
// need to observe storage directly rather than through the port.
func (s *Store) DB() *sql.DB {
	return s.db
}

// CreateRepo implements ingestion.Store.
func (s *Store) CreateRepo(ctx context.Context, repo ingestion.NewRepo) (ingestion.Repo, error) {
	encrypted, err := crypto.Encrypt(s.key, []byte(repo.Token))
	if err != nil {
		return ingestion.Repo{}, fmt.Errorf("encrypt token: %w", err)
	}

	webhookSecret, err := crypto.RandomHex(32)
	if err != nil {
		return ingestion.Repo{}, fmt.Errorf("generate webhook secret: %w", err)
	}

	repoRecord := ingestion.Repo{
		ID:                 uuid.NewString(),
		Forge:              repo.Forge,
		Identifier:         repo.Identifier,
		ForgejoInstanceURL: repo.ForgejoInstanceURL,
		TokenMasked:        ingestion.MaskToken(repo.Token),
		WebhookSecret:      webhookSecret,
		IngestionStatus:    ingestion.StatusPending,
		CreatedAt:          time.Now().UTC(),
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO repos (id, forge, identifier, forgejo_instance_url, token_encrypted, token_masked, webhook_secret, ingestion_status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		repoRecord.ID, string(repoRecord.Forge), repoRecord.Identifier, nullable(repoRecord.ForgejoInstanceURL),
		encrypted, repoRecord.TokenMasked, repoRecord.WebhookSecret, string(repoRecord.IngestionStatus), repoRecord.CreatedAt,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return ingestion.Repo{}, ingestion.ErrRepoAlreadyTracked
		}

		return ingestion.Repo{}, fmt.Errorf("insert repo: %w", err)
	}

	return repoRecord, nil
}

// isUniqueConstraintErr reports whether err is a SQLite UNIQUE constraint
// violation -- detected via the driver's own error code
// (SQLITE_CONSTRAINT_UNIQUE), not by matching the error message string.
func isUniqueConstraintErr(err error) bool {
	var sqliteErr *sqlitedriver.Error

	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}

// ListRepoIdentifiers implements ingestion.Store.
func (s *Store) ListRepoIdentifiers(ctx context.Context, forge ingestion.Forge, instanceURL string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT identifier FROM repos
		WHERE forge = ? AND forgejo_instance_url IS ?
	`, string(forge), nullable(instanceURL))
	if err != nil {
		return nil, fmt.Errorf("query repo identifiers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var identifiers []string

	for rows.Next() {
		var identifier string
		if err := rows.Scan(&identifier); err != nil {
			return nil, fmt.Errorf("scan repo identifier: %w", err)
		}

		identifiers = append(identifiers, identifier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repo identifiers: %w", err)
	}

	return identifiers, nil
}

// ListRepos implements ingestion.Store.
func (s *Store) ListRepos(ctx context.Context, filter ingestion.RepoListFilter) ([]ingestion.Repo, bool, error) {
	query := `
		SELECT id, forge, identifier, forgejo_instance_url, token_masked, webhook_secret, ingestion_status, ingestion_status_reason, reconcile_etag, created_at
		FROM repos
	`
	args := []any{}

	if filter.Forge != "" {
		query += " WHERE forge = ?"
		args = append(args, string(filter.Forge))
	}

	query += " ORDER BY created_at, id"

	// Fetching one extra row is what tells the caller whether a next page
	// exists, without a separate COUNT(*) round trip.
	if filter.Limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, filter.Limit+1, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query repos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var repos []ingestion.Repo

	for rows.Next() {
		repo, err := scanRepo(rows)
		if err != nil {
			return nil, false, err
		}

		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate repos: %w", err)
	}

	hasMore := filter.Limit > 0 && len(repos) > filter.Limit
	if hasMore {
		repos = repos[:filter.Limit]
	}

	return repos, hasMore, nil
}

// ErrRepoNotFound is returned when no repo matches the requested id.
var ErrRepoNotFound = errors.New("repo not found")

// GetRepo implements ingestion.Store.
func (s *Store) GetRepo(ctx context.Context, id string) (ingestion.Repo, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, forge, identifier, forgejo_instance_url, token_masked, webhook_secret, ingestion_status, ingestion_status_reason, reconcile_etag, created_at
		FROM repos WHERE id = ?
	`, id)

	repo, err := scanRepo(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ingestion.Repo{}, ErrRepoNotFound
	}

	if err != nil {
		return ingestion.Repo{}, err
	}

	return repo, nil
}

// LocateJob implements ingestion.JobLocator. A job id that belongs to a
// different run is not found: the pair must match.
func (s *Store) LocateJob(ctx context.Context, runID, jobID string) (ingestion.JobLocation, error) {
	var location ingestion.JobLocation

	err := s.db.QueryRowContext(ctx, `
		SELECT r.repo_id, j.forge_job_id, j.forge_url
		FROM jobs j
		JOIN runs r ON r.id = j.run_id
		WHERE j.id = ? AND j.run_id = ?
	`, jobID, runID).Scan(&location.RepoID, &location.ForgeJobID, &location.ForgeURL)
	if errors.Is(err, sql.ErrNoRows) {
		return ingestion.JobLocation{}, ingestion.ErrJobNotFound
	}

	if err != nil {
		return ingestion.JobLocation{}, fmt.Errorf("locate job: %w", err)
	}

	return location, nil
}

// DeleteRepo implements ingestion.Store. The schema's ON DELETE CASCADE
// takes care of its runs, jobs, and steps.
func (s *Store) DeleteRepo(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM repos WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete repo: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected: %w", err)
	}

	if rows == 0 {
		return ErrRepoNotFound
	}

	return nil
}

// SetReconcileETag implements ingestion.Store.
func (s *Store) SetReconcileETag(ctx context.Context, id string, etag string) error {
	result, err := s.db.ExecContext(ctx, "UPDATE repos SET reconcile_etag = ? WHERE id = ?", etag, id)
	if err != nil {
		return fmt.Errorf("update reconcile etag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected: %w", err)
	}

	if rows == 0 {
		return ErrRepoNotFound
	}

	return nil
}

// SetIngestionStatus implements ingestion.Store.
func (s *Store) SetIngestionStatus(ctx context.Context, id string, status ingestion.Status, reason string) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE repos SET ingestion_status = ?, ingestion_status_reason = ? WHERE id = ?",
		string(status), nullable(reason), id,
	)
	if err != nil {
		return fmt.Errorf("update ingestion status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected: %w", err)
	}

	if rows == 0 {
		return ErrRepoNotFound
	}

	return nil
}

// RepoToken implements ingestion.Store.
func (s *Store) RepoToken(ctx context.Context, id string) (string, error) {
	var encrypted []byte

	err := s.db.QueryRowContext(ctx, "SELECT token_encrypted FROM repos WHERE id = ?", id).Scan(&encrypted)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrRepoNotFound
	}

	if err != nil {
		return "", fmt.Errorf("query token: %w", err)
	}

	plaintext, err := crypto.Decrypt(s.key, encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt token: %w", err)
	}

	return string(plaintext), nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRepo(row rowScanner) (ingestion.Repo, error) {
	var (
		repo               ingestion.Repo
		forge              string
		status             string
		forgejoInstanceURL sql.NullString
		statusReason       sql.NullString
	)

	err := row.Scan(
		&repo.ID, &forge, &repo.Identifier, &forgejoInstanceURL, &repo.TokenMasked,
		&repo.WebhookSecret, &status, &statusReason, &repo.ReconcileETag, &repo.CreatedAt,
	)
	if err != nil {
		return ingestion.Repo{}, fmt.Errorf("scan repo: %w", err)
	}

	repo.Forge = ingestion.Forge(forge)
	repo.IngestionStatus = ingestion.Status(status)
	repo.ForgejoInstanceURL = forgejoInstanceURL.String
	repo.IngestionStatusReason = statusReason.String

	return repo, nil
}

func nullable(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

var _ ingestion.ForgeTokenStore = (*Store)(nil)

// tokenScope normalizes the (forge, instance) a token is saved under: GitHub
// has no instance, and a trailing slash doesn't make a Forgejo instance
// another one.
func tokenScope(forge ingestion.Forge, instanceURL string) string {
	if forge != ingestion.ForgeForgejo {
		return ""
	}

	return strings.TrimRight(strings.TrimSpace(instanceURL), "/")
}

// SaveForgeToken implements ingestion.ForgeTokenStore.
func (s *Store) SaveForgeToken(ctx context.Context, forge ingestion.Forge, instanceURL, token string) (ingestion.SavedToken, error) {
	encrypted, err := crypto.Encrypt(s.key, []byte(token))
	if err != nil {
		return ingestion.SavedToken{}, fmt.Errorf("encrypt token: %w", err)
	}

	scope := tokenScope(forge, instanceURL)
	saved := ingestion.SavedToken{
		ID:                 uuid.NewString(),
		Forge:              forge,
		ForgejoInstanceURL: scope,
		TokenMasked:        ingestion.MaskToken(token),
	}

	// Replacing keeps the row's id, so a client holding it can still delete
	// it. RETURNING gives back the id that is actually stored.
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO forge_tokens (id, forge, instance_url, token_encrypted, token_masked, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (forge, instance_url) DO UPDATE SET
			token_encrypted = excluded.token_encrypted,
			token_masked = excluded.token_masked,
			created_at = excluded.created_at
		RETURNING id
	`, saved.ID, string(forge), scope, encrypted, saved.TokenMasked, time.Now().UTC()).Scan(&saved.ID)
	if err != nil {
		return ingestion.SavedToken{}, fmt.Errorf("save forge token: %w", err)
	}

	return saved, nil
}

// ListForgeTokens implements ingestion.ForgeTokenStore.
func (s *Store) ListForgeTokens(ctx context.Context) ([]ingestion.SavedToken, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, forge, instance_url, token_masked FROM forge_tokens ORDER BY forge, instance_url")
	if err != nil {
		return nil, fmt.Errorf("query forge tokens: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tokens []ingestion.SavedToken

	for rows.Next() {
		var (
			t     ingestion.SavedToken
			forge string
		)

		if err := rows.Scan(&t.ID, &forge, &t.ForgejoInstanceURL, &t.TokenMasked); err != nil {
			return nil, fmt.Errorf("scan forge token: %w", err)
		}

		t.Forge = ingestion.Forge(forge)
		tokens = append(tokens, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate forge tokens: %w", err)
	}

	return tokens, nil
}

// DeleteForgeToken implements ingestion.ForgeTokenStore.
func (s *Store) DeleteForgeToken(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM forge_tokens WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete forge token: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete forge token: %w", err)
	}

	if n == 0 {
		return ingestion.ErrSavedTokenNotFound
	}

	return nil
}

// ForgeToken implements ingestion.ForgeTokenStore.
func (s *Store) ForgeToken(ctx context.Context, forge ingestion.Forge, instanceURL string) (string, error) {
	var encrypted []byte

	err := s.db.QueryRowContext(ctx,
		"SELECT token_encrypted FROM forge_tokens WHERE forge = ? AND instance_url = ?",
		string(forge), tokenScope(forge, instanceURL),
	).Scan(&encrypted)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ingestion.ErrSavedTokenNotFound
	}

	if err != nil {
		return "", fmt.Errorf("query forge token: %w", err)
	}

	plaintext, err := crypto.Decrypt(s.key, encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt forge token: %w", err)
	}

	return string(plaintext), nil
}

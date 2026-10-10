package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

// ActiveQuarantines implements metrics.Store. A mark whose expiry has passed is
// left in the table and ignored here; the next QuarantineStep replaces it.
func (s *Store) ActiveQuarantines(ctx context.Context, now time.Time) (map[metrics.QuarantineKey]metrics.Quarantine, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT repo_id, pipeline_name, step_name, note, quarantined_at, expires_at
		FROM step_quarantines
		WHERE expires_at > ?
	`, now.UTC())
	if err != nil {
		return nil, fmt.Errorf("query quarantines: %w", err)
	}
	defer func() { _ = rows.Close() }()

	active := make(map[metrics.QuarantineKey]metrics.Quarantine)

	for rows.Next() {
		var (
			key metrics.QuarantineKey
			q   metrics.Quarantine
		)

		if err := rows.Scan(&key.Pipeline.RepoID, &key.Pipeline.Name, &key.Step, &q.Note, &q.QuarantinedAt, &q.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan quarantine: %w", err)
		}

		active[key] = q
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read quarantines: %w", err)
	}

	return active, nil
}

// QuarantineStep implements metrics.Store.
func (s *Store) QuarantineStep(ctx context.Context, key metrics.QuarantineKey, note string, now time.Time) (metrics.Quarantine, error) {
	q := metrics.Quarantine{Note: note, QuarantinedAt: now.UTC(), ExpiresAt: now.UTC().Add(metrics.QuarantineDuration)}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO step_quarantines (repo_id, pipeline_name, step_name, note, quarantined_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (repo_id, pipeline_name, step_name)
		DO UPDATE SET note = excluded.note, quarantined_at = excluded.quarantined_at, expires_at = excluded.expires_at
	`, key.Pipeline.RepoID, key.Pipeline.Name, key.Step, q.Note, q.QuarantinedAt, q.ExpiresAt)
	if err != nil {
		return metrics.Quarantine{}, fmt.Errorf("quarantine step: %w", err)
	}

	return q, nil
}

// UnquarantineStep implements metrics.Store.
func (s *Store) UnquarantineStep(ctx context.Context, key metrics.QuarantineKey) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM step_quarantines WHERE repo_id = ? AND pipeline_name = ? AND step_name = ?`,
		key.Pipeline.RepoID, key.Pipeline.Name, key.Step)
	if err != nil {
		return fmt.Errorf("un-quarantine step: %w", err)
	}

	return nil
}

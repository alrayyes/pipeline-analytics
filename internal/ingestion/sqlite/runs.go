package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/google/uuid"
)

// RepoByIdentifier implements ingestion.RunStore.
func (s *Store) RepoByIdentifier(ctx context.Context, forge ingestion.Forge, identifier string) (ingestion.Repo, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, forge, identifier, forgejo_instance_url, token_masked, webhook_secret, ingestion_status, ingestion_status_reason, reconcile_etag, created_at
		FROM repos WHERE forge = ? AND identifier = ?
	`, string(forge), identifier)

	repo, err := scanRepo(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ingestion.Repo{}, ErrRepoNotFound
	}

	if err != nil {
		return ingestion.Repo{}, err
	}

	return repo, nil
}

// RunStates implements ingestion.RunStore.
func (s *Store) RunStates(ctx context.Context, repoID string) (map[string]ingestion.RunState, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT forge_run_id, status, conclusion FROM runs WHERE repo_id = ?", repoID)
	if err != nil {
		return nil, fmt.Errorf("query run states: %w", err)
	}
	defer func() { _ = rows.Close() }()

	states := make(map[string]ingestion.RunState)

	for rows.Next() {
		var (
			forgeRunID string
			state      ingestion.RunState
			conclusion sql.NullString
		)

		if err := rows.Scan(&forgeRunID, &state.Status, &conclusion); err != nil {
			return nil, fmt.Errorf("scan run state: %w", err)
		}

		state.Conclusion = conclusion.String
		states[forgeRunID] = state
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run states: %w", err)
	}

	return states, nil
}

// UpsertRun implements ingestion.RunStore.
func (s *Store) UpsertRun(ctx context.Context, run ingestion.Run) (ingestion.Run, error) {
	existingID, err := s.findID(ctx, "SELECT id FROM runs WHERE repo_id = ? AND forge_run_id = ?", run.RepoID, run.ForgeRunID)
	if err != nil {
		return ingestion.Run{}, err
	}

	if existingID != "" {
		run.ID = existingID

		_, err := s.db.ExecContext(ctx, `
			UPDATE runs SET pipeline_name = ?, status = ?, conclusion = ?, started_at = ?, completed_at = ?, forge_url = ?
			WHERE id = ?
		`, run.PipelineName, run.Status, nullable(run.Conclusion), run.StartedAt, run.CompletedAt, run.ForgeURL, run.ID)
		if err != nil {
			return ingestion.Run{}, fmt.Errorf("update run: %w", err)
		}

		return run, nil
	}

	run.ID = uuid.NewString()

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO runs (id, repo_id, forge_run_id, pipeline_name, status, conclusion, started_at, completed_at, forge_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, run.ID, run.RepoID, run.ForgeRunID, run.PipelineName, run.Status, nullable(run.Conclusion), run.StartedAt, run.CompletedAt, run.ForgeURL)
	if err != nil {
		return ingestion.Run{}, fmt.Errorf("insert run: %w", err)
	}

	return run, nil
}

// UpsertJob implements ingestion.RunStore.
func (s *Store) UpsertJob(ctx context.Context, job ingestion.Job) (ingestion.Job, error) {
	existingID, err := s.findID(ctx, "SELECT id FROM jobs WHERE run_id = ? AND forge_job_id = ?", job.RunID, job.ForgeJobID)
	if err != nil {
		return ingestion.Job{}, err
	}

	if existingID != "" {
		job.ID = existingID

		_, err := s.db.ExecContext(ctx, `
			UPDATE jobs SET name = ?, status = ?, conclusion = ?, queued_at = ?, started_at = ?, completed_at = ?, forge_url = ?
			WHERE id = ?
		`, job.Name, job.Status, nullable(job.Conclusion), job.QueuedAt, job.StartedAt, job.CompletedAt, job.ForgeURL, job.ID)
		if err != nil {
			return ingestion.Job{}, fmt.Errorf("update job: %w", err)
		}

		return job, nil
	}

	job.ID = uuid.NewString()

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO jobs (id, run_id, forge_job_id, name, status, conclusion, queued_at, started_at, completed_at, forge_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.RunID, job.ForgeJobID, job.Name, job.Status, nullable(job.Conclusion), job.QueuedAt, job.StartedAt, job.CompletedAt, job.ForgeURL)
	if err != nil {
		return ingestion.Job{}, fmt.Errorf("insert job: %w", err)
	}

	return job, nil
}

// ReplaceSteps implements ingestion.RunStore.
func (s *Store) ReplaceSteps(ctx context.Context, jobID string, steps []ingestion.Step) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "DELETE FROM steps WHERE job_id = ?", jobID); err != nil {
		return fmt.Errorf("clear steps: %w", err)
	}

	for _, step := range steps {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO steps (id, job_id, number, name, status, conclusion, started_at, completed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, uuid.NewString(), jobID, step.Number, step.Name, step.Status, nullable(step.Conclusion), step.StartedAt, step.CompletedAt)
		if err != nil {
			return fmt.Errorf("insert step %d: %w", step.Number, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit steps: %w", err)
	}

	return nil
}

// findID returns the id column from the first row query+args matches, or ""
// if there's no match.
func (s *Store) findID(ctx context.Context, query string, args ...any) (string, error) {
	var id string

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("find id: %w", err)
	}

	return id, nil
}

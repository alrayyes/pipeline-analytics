// Package sqlite is the SQLite-backed adapter for the metrics.Store port,
// reading the same repos/runs/jobs/steps tables ingestion writes to.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

// Store implements metrics.Store against a SQLite database.
type Store struct {
	db *sql.DB
}

// NewStore returns a Store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// ListPipelines implements metrics.Store.
func (s *Store) ListPipelines(ctx context.Context) ([]metrics.PipelineRef, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT repo_id, pipeline_name FROM runs")
	if err != nil {
		return nil, fmt.Errorf("query pipelines: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var refs []metrics.PipelineRef

	for rows.Next() {
		var ref metrics.PipelineRef
		if err := rows.Scan(&ref.RepoID, &ref.Name); err != nil {
			return nil, fmt.Errorf("scan pipeline: %w", err)
		}

		refs = append(refs, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pipelines: %w", err)
	}

	return refs, nil
}

// PipelineRuns implements metrics.Store.
func (s *Store) PipelineRuns(ctx context.Context, ref metrics.PipelineRef, window metrics.Window) ([]metrics.RunRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, status, conclusion, started_at, completed_at
		FROM runs
		WHERE repo_id = ? AND pipeline_name = ?
		ORDER BY started_at DESC, id DESC
		LIMIT ?
	`, ref.RepoID, ref.Name, window.RunCount)
	if err != nil {
		return nil, fmt.Errorf("query pipeline runs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var runs []metrics.RunRecord

	for rows.Next() {
		var (
			run        metrics.RunRecord
			conclusion sql.NullString
		)

		if err := rows.Scan(&run.ID, &run.Status, &conclusion, &run.StartedAt, &run.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}

		run.Conclusion = conclusion.String
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pipeline runs: %w", err)
	}

	return runs, nil
}

// PipelineSteps implements metrics.Store.
func (s *Store) PipelineSteps(ctx context.Context, ref metrics.PipelineRef, window metrics.Window) ([]metrics.StepOccurrence, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.name, s.status, s.conclusion, s.started_at, s.completed_at, j.queued_at, j.started_at, j.forge_url
		FROM steps s
		JOIN jobs j ON j.id = s.job_id
		WHERE j.run_id IN (
			SELECT id FROM runs WHERE repo_id = ? AND pipeline_name = ? ORDER BY started_at DESC, id DESC LIMIT ?
		)
	`, ref.RepoID, ref.Name, window.RunCount)
	if err != nil {
		return nil, fmt.Errorf("query pipeline steps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var occurrences []metrics.StepOccurrence

	for rows.Next() {
		var (
			occ        metrics.StepOccurrence
			conclusion sql.NullString
		)

		err := rows.Scan(
			&occ.Name, &occ.Status, &conclusion, &occ.StartedAt, &occ.CompletedAt,
			&occ.JobQueuedAt, &occ.JobStartedAt, &occ.JobForgeURL,
		)
		if err != nil {
			return nil, fmt.Errorf("scan step: %w", err)
		}

		occ.Conclusion = conclusion.String
		occurrences = append(occurrences, occ)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pipeline steps: %w", err)
	}

	return occurrences, nil
}

// RepoUsage implements metrics.Store. Each tracked pipeline within the repo
// contributes its own window's worth of runs, so a rarely run workflow
// isn't crowded out of the window by a busy one.
func (s *Store) RepoUsage(ctx context.Context, repoID string, window metrics.Window) ([]metrics.UsageRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.pipeline_name, j.started_at, j.completed_at
		FROM jobs j
		JOIN runs r ON r.id = j.run_id
		WHERE r.id IN (
			SELECT id FROM (
				SELECT id, ROW_NUMBER() OVER (PARTITION BY pipeline_name ORDER BY started_at DESC, id DESC) AS rn
				FROM runs WHERE repo_id = ?
			) WHERE rn <= ?
		)
	`, repoID, window.RunCount)
	if err != nil {
		return nil, fmt.Errorf("query repo usage: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var records []metrics.UsageRecord

	for rows.Next() {
		var (
			pipelineName           string
			startedAt, completedAt sql.NullTime
		)

		if err := rows.Scan(&pipelineName, &startedAt, &completedAt); err != nil {
			return nil, fmt.Errorf("scan usage job: %w", err)
		}

		if !startedAt.Valid || !completedAt.Valid {
			continue
		}

		records = append(records, metrics.UsageRecord{
			PipelineName: pipelineName,
			ExecSeconds:  completedAt.Time.Sub(startedAt.Time).Seconds(),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repo usage: %w", err)
	}

	return records, nil
}

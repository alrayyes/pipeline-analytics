// Package sqlite is the SQLite-backed adapter for the metrics.Store port,
// reading the same repos/runs/jobs/steps tables ingestion writes to.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
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

// pipelineListQuery builds the SQL and args ListPipelines runs for filter.
func pipelineListQuery(filter metrics.PipelineListFilter) (string, []any) {
	query := "SELECT DISTINCT r.repo_id, r.pipeline_name FROM runs r"
	args := []any{}

	// The runs table has no forge column of its own, so filtering by forge
	// needs a join to the owning repo -- skipped when Forge is unset, to
	// keep the common (unfiltered) query simple.
	if filter.Forge != "" {
		query += " JOIN repos p ON p.id = r.repo_id"
	}

	var where string

	if filter.RepoID != "" {
		where += " r.repo_id = ?"
		args = append(args, filter.RepoID)
	}

	if filter.Forge != "" {
		if where != "" {
			where += " AND"
		}

		where += " p.forge = ?"
		args = append(args, filter.Forge)
	}

	if where != "" {
		query += " WHERE"
		query += where
	}

	query += " ORDER BY r.repo_id, r.pipeline_name"

	// Fetching one extra row is what tells the caller whether a next page
	// exists, without a separate COUNT(*) round trip.
	if filter.Limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, filter.Limit+1, filter.Offset)
	}

	return query, args
}

// ListPipelines implements metrics.Store.
func (s *Store) ListPipelines(ctx context.Context, filter metrics.PipelineListFilter) ([]metrics.PipelineRef, bool, error) {
	query, args := pipelineListQuery(filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query pipelines: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var refs []metrics.PipelineRef

	for rows.Next() {
		var ref metrics.PipelineRef
		if err := rows.Scan(&ref.RepoID, &ref.Name); err != nil {
			return nil, false, fmt.Errorf("scan pipeline: %w", err)
		}

		refs = append(refs, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate pipelines: %w", err)
	}

	hasMore := filter.Limit > 0 && len(refs) > filter.Limit
	if hasMore {
		refs = refs[:filter.Limit]
	}

	return refs, hasMore, nil
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
		SELECT s.name, s.status, s.conclusion, s.started_at, s.completed_at, j.queued_at, j.started_at, j.forge_url, j.run_id, r.started_at, j.id
		FROM steps s
		JOIN jobs j ON j.id = s.job_id
		JOIN runs r ON r.id = j.run_id
		WHERE j.run_id IN (
			SELECT id FROM runs WHERE repo_id = ? AND pipeline_name = ? ORDER BY started_at DESC, id DESC LIMIT ?
		)
	`, ref.RepoID, ref.Name, window.RunCount)
	if err != nil {
		return nil, fmt.Errorf("query pipeline steps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanStepOccurrences(rows)
}

// WindowRuns implements metrics.Store.
func (s *Store) WindowRuns(ctx context.Context, filter metrics.RunWindowFilter) ([]metrics.WindowRun, error) {
	query := "SELECT r.id, r.repo_id, r.pipeline_name, r.status, r.conclusion, r.started_at, r.completed_at FROM runs r"
	args := []any{}

	// As in pipelineListQuery: runs has no forge column, so the join to the
	// owning repo is only added when a forge filter needs it.
	if filter.Forge != "" {
		query += " JOIN repos p ON p.id = r.repo_id"
	}

	query += " WHERE r.started_at >= ? AND r.started_at < ?"
	args = append(args, filter.Since, filter.Until)

	if filter.RepoID != "" {
		query += " AND r.repo_id = ?"
		args = append(args, filter.RepoID)
	}

	if filter.Forge != "" {
		query += " AND p.forge = ?"
		args = append(args, filter.Forge)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query window runs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var runs []metrics.WindowRun

	for rows.Next() {
		var (
			wr         metrics.WindowRun
			conclusion sql.NullString
		)

		err := rows.Scan(&wr.Run.ID, &wr.Pipeline.RepoID, &wr.Pipeline.Name, &wr.Run.Status, &conclusion, &wr.Run.StartedAt, &wr.Run.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("scan window run: %w", err)
		}

		wr.Run.Conclusion = conclusion.String
		runs = append(runs, wr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate window runs: %w", err)
	}

	return runs, nil
}

// WindowSteps implements metrics.Store.
func (s *Store) WindowSteps(ctx context.Context, filter metrics.RunWindowFilter) ([]metrics.WindowStep, error) {
	query := `
		SELECT r.repo_id, r.pipeline_name,
			s.name, s.status, s.conclusion, s.started_at, s.completed_at, j.queued_at, j.started_at, j.forge_url, j.run_id, r.started_at
		FROM steps s
		JOIN jobs j ON j.id = s.job_id
		JOIN runs r ON r.id = j.run_id`
	args := []any{}

	if filter.Forge != "" {
		query += " JOIN repos p ON p.id = r.repo_id"
	}

	query += " WHERE r.started_at >= ? AND r.started_at < ?"
	args = append(args, filter.Since, filter.Until)

	if filter.RepoID != "" {
		query += " AND r.repo_id = ?"
		args = append(args, filter.RepoID)
	}

	if filter.Forge != "" {
		query += " AND p.forge = ?"
		args = append(args, filter.Forge)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query window steps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var steps []metrics.WindowStep

	for rows.Next() {
		var (
			ws         metrics.WindowStep
			conclusion sql.NullString
		)

		err := rows.Scan(
			&ws.Pipeline.RepoID, &ws.Pipeline.Name,
			&ws.Step.Name, &ws.Step.Status, &conclusion, &ws.Step.StartedAt, &ws.Step.CompletedAt,
			&ws.Step.JobQueuedAt, &ws.Step.JobStartedAt, &ws.Step.JobForgeURL, &ws.Step.RunID, &ws.Step.RunStartedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan window step: %w", err)
		}

		ws.Step.Conclusion = conclusion.String
		steps = append(steps, ws)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate window steps: %w", err)
	}

	return steps, nil
}

// runListQuery builds the SQL and args ListRuns runs for filter.
func runListQuery(filter metrics.RunListFilter) (string, []any) {
	query := `
		SELECT r.id, r.repo_id, r.pipeline_name, r.status, r.conclusion, r.started_at, r.completed_at,
			r.branch, r.head_sha, r.head_message, r.actor, r.forge_url
		FROM runs r`
	args := []any{}

	if filter.Forge != "" {
		query += " JOIN repos p ON p.id = r.repo_id"
	}

	query += " WHERE 1 = 1"

	if filter.RepoID != "" {
		query += " AND r.repo_id = ?"
		args = append(args, filter.RepoID)
	}

	if filter.Forge != "" {
		query += " AND p.forge = ?"
		args = append(args, filter.Forge)
	}

	switch filter.Status {
	case metrics.RunStatusFailed:
		query += " AND r.conclusion IN ('failure', 'timed_out')"
	case metrics.RunStatusRunning:
		query += " AND r.status != 'completed'"
	case metrics.RunStatusSuccess:
		query += " AND r.conclusion = 'success'"
	case metrics.RunStatusAll:
	}

	// A run that hasn't started has no start time yet and is the newest, so
	// NULLs sort first.
	query += " ORDER BY r.started_at IS NULL DESC, r.started_at DESC, r.id DESC"

	// One extra row tells the caller whether another page exists without a
	// separate COUNT(*), as in pipelineListQuery.
	if filter.Limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, filter.Limit+1, filter.Offset)
	}

	return query, args
}

// ListRuns implements metrics.Store.
func (s *Store) ListRuns(ctx context.Context, filter metrics.RunListFilter) ([]metrics.RunEntry, bool, error) {
	query, args := runListQuery(filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query runs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var runs []metrics.RunEntry

	for rows.Next() {
		run, err := scanRunEntry(rows)
		if err != nil {
			return nil, false, err
		}

		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate runs: %w", err)
	}

	hasMore := filter.Limit > 0 && len(runs) > filter.Limit
	if hasMore {
		runs = runs[:filter.Limit]
	}

	if err := s.attachSteps(ctx, runs); err != nil {
		return nil, false, err
	}

	return runs, hasMore, nil
}

func scanRunEntry(rows *sql.Rows) (metrics.RunEntry, error) {
	var (
		run                                     metrics.RunEntry
		conclusion, branch, sha, message, actor sql.NullString
	)

	err := rows.Scan(
		&run.ID, &run.Pipeline.RepoID, &run.Pipeline.Name, &run.Status, &conclusion, &run.StartedAt, &run.CompletedAt,
		&branch, &sha, &message, &actor, &run.ForgeURL,
	)
	if err != nil {
		return metrics.RunEntry{}, fmt.Errorf("scan run: %w", err)
	}

	run.Conclusion = conclusion.String
	run.Branch, run.SHA, run.Message, run.Actor = branch.String, sha.String, message.String, actor.String
	run.Steps = []metrics.RunStep{}

	return run, nil
}

// attachSteps fills each run's Steps with one query for the whole page,
// rather than one per run.
func (s *Store) attachSteps(ctx context.Context, runs []metrics.RunEntry) error {
	if len(runs) == 0 {
		return nil
	}

	index := make(map[string]int, len(runs))
	ids := make([]string, 0, len(runs))

	for i, run := range runs {
		index[run.ID] = i
		ids = append(ids, run.ID)
	}

	// The ids go in as one JSON array read by json_each, so the query text
	// is fixed however many runs are on the page -- no SQL assembled from
	// a variable-length placeholder list.
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		return fmt.Errorf("encode run ids: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT j.run_id, s.name, s.status, s.conclusion, j.forge_url, j.id
		FROM steps s
		JOIN jobs j ON j.id = s.job_id
		WHERE j.run_id IN (SELECT value FROM json_each(?))
		ORDER BY j.run_id, j.started_at, s.number
	`, string(idsJSON))
	if err != nil {
		return fmt.Errorf("query run steps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			runID      string
			step       metrics.RunStep
			conclusion sql.NullString
		)

		if err := rows.Scan(&runID, &step.Name, &step.Status, &conclusion, &step.ForgeURL, &step.JobID); err != nil {
			return fmt.Errorf("scan run step: %w", err)
		}

		step.Conclusion = conclusion.String
		runs[index[runID]].Steps = append(runs[index[runID]].Steps, step)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate run steps: %w", err)
	}

	return nil
}

// RunSteps implements metrics.Store.
func (s *Store) RunSteps(ctx context.Context, runID string) ([]metrics.StepOccurrence, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.name, s.status, s.conclusion, s.started_at, s.completed_at, j.queued_at, j.started_at, j.forge_url, j.run_id, r.started_at, j.id
		FROM steps s
		JOIN jobs j ON j.id = s.job_id
		JOIN runs r ON r.id = j.run_id
		WHERE j.run_id = ?
		ORDER BY j.started_at, s.number
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("query run steps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanStepOccurrences(rows)
}

func scanStepOccurrences(rows *sql.Rows) ([]metrics.StepOccurrence, error) {
	var occurrences []metrics.StepOccurrence

	for rows.Next() {
		var (
			occ        metrics.StepOccurrence
			conclusion sql.NullString
		)

		err := rows.Scan(
			&occ.Name, &occ.Status, &conclusion, &occ.StartedAt, &occ.CompletedAt,
			&occ.JobQueuedAt, &occ.JobStartedAt, &occ.JobForgeURL, &occ.RunID, &occ.RunStartedAt, &occ.JobID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan step: %w", err)
		}

		occ.Conclusion = conclusion.String
		occurrences = append(occurrences, occ)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate steps: %w", err)
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

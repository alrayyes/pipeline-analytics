package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RunStatus is a status bucket the run list can be filtered to, matching the
// API's RunStatusFilter.
type RunStatus string

// The run-list status buckets. Failed is a concluded failure or timeout,
// Running is anything not yet completed (queued or in progress), and Success
// is a concluded success; a cancelled or skipped run is in All only.
const (
	RunStatusAll     RunStatus = "all"
	RunStatusFailed  RunStatus = "failed"
	RunStatusRunning RunStatus = "running"
	RunStatusSuccess RunStatus = "success"
)

// ErrInvalidRunStatus is returned by ParseRunStatus for a value outside the
// documented set.
var ErrInvalidRunStatus = errors.New("invalid run status")

// ParseRunStatus interprets the API's status query parameter. Empty means
// all, the contract's default; anything else outside the set is an error
// rather than a silent fallback, so a typo can't look like "no failures".
func ParseRunStatus(raw string) (RunStatus, error) {
	switch status := RunStatus(raw); status {
	case "":
		return RunStatusAll, nil
	case RunStatusAll, RunStatusFailed, RunStatusRunning, RunStatusSuccess:
		return status, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidRunStatus, raw)
	}
}

// RunListFilter narrows ListRuns to one repo, forge and status bucket, and to
// one page. A zero Limit means unlimited.
type RunListFilter struct {
	RepoID string
	Forge  string
	Branch string
	Status RunStatus
	Limit  int
	Offset int
}

// RunEntry is one run in the run list: its pipeline, commit metadata (empty
// where the forge or an older row lacks it), forge link and steps.
type RunEntry struct {
	ID          string
	Pipeline    PipelineRef
	Status      string
	Conclusion  string
	StartedAt   *time.Time
	CompletedAt *time.Time
	Branch      string
	SHA         string
	Message     string
	Actor       string
	ForgeURL    string
	// Steps are in recorded order (job, then step number). Never nil.
	Steps []RunStep
}

// DurationSeconds is the run's wall-clock span. It reports false for a run
// that is still going or never started, rather than a misleading zero.
func (e RunEntry) DurationSeconds() (float64, bool) {
	return durationSeconds(e.StartedAt, e.CompletedAt)
}

// ListRuns returns a page of runs, newest first, and whether more match.
func (s *Service) ListRuns(ctx context.Context, filter RunListFilter) ([]RunEntry, bool, error) {
	runs, hasMore, err := s.store.ListRuns(ctx, filter)
	if err != nil {
		return nil, false, fmt.Errorf("list runs: %w", err)
	}

	return runs, hasMore, nil
}

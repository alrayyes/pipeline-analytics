package ingestion

import (
	"context"
	"errors"
	"fmt"
)

// ErrJobNotFound is returned when a run and job id name no recorded job.
var ErrJobNotFound = errors.New("job not found")

const (
	// DefaultLogLines is how many lines a log request gets when it names none.
	DefaultLogLines = 200
	// MaxLogLines is the most lines one request may ask for.
	MaxLogLines = 1000
)

// JobLogReason is why a job's log wasn't returned, as the API reports it.
type JobLogReason string

// The reasons a log can be unavailable.
const (
	LogUnsupported JobLogReason = "unsupported"
	LogExpired     JobLogReason = "expired"
	LogForbidden   JobLogReason = "forbidden"
	LogUnreachable JobLogReason = "unreachable"
)

// JobLocation is where a recorded job lives on its forge.
type JobLocation struct {
	RepoID     string
	ForgeJobID string
	ForgeURL   string
}

// JobLocator resolves a run's job to its location, or ErrJobNotFound.
type JobLocator interface {
	LocateJob(ctx context.Context, runID, jobID string) (JobLocation, error)
}

// JobLogRepos is the slice of the repo store the log service needs.
type JobLogRepos interface {
	GetRepo(ctx context.Context, id string) (Repo, error)
	RepoToken(ctx context.Context, id string) (string, error)
}

// JobLogResult is a job's log tail, or why there isn't one. ForgeURL is
// always set, so a caller can still link out.
type JobLogResult struct {
	Available bool
	Reason    JobLogReason
	Lines     []string
	Truncated bool
	ForgeURL  string
}

// JobLogService reads a job's log from its forge on demand. Nothing is
// stored (openspec/changes/api-job-log).
type JobLogService struct {
	locator JobLocator
	repos   JobLogRepos
	readers map[Forge]JobLogReader
}

// NewJobLogService returns a JobLogService. A forge with no reader in readers
// is reported as unsupported.
func NewJobLogService(locator JobLocator, repos JobLogRepos, readers map[Forge]JobLogReader) *JobLogService {
	return &JobLogService{locator: locator, repos: repos, readers: readers}
}

// Tail returns the last lines of a job's log. Lines is clamped to
// [1, MaxLogLines], defaulting to DefaultLogLines. A log the forge can't
// supply is a result with a Reason; only an unknown job or an unexpected
// failure is an error.
func (s *JobLogService) Tail(ctx context.Context, runID, jobID string, lines int) (JobLogResult, error) {
	if lines <= 0 {
		lines = DefaultLogLines
	}

	lines = min(lines, MaxLogLines)

	location, err := s.locator.LocateJob(ctx, runID, jobID)
	if err != nil {
		return JobLogResult{}, fmt.Errorf("locate job: %w", err)
	}

	unavailable := func(reason JobLogReason) JobLogResult {
		return JobLogResult{Reason: reason, ForgeURL: location.ForgeURL}
	}

	repo, err := s.repos.GetRepo(ctx, location.RepoID)
	if err != nil {
		return JobLogResult{}, fmt.Errorf("get repo: %w", err)
	}

	reader, ok := s.readers[repo.Forge]
	if !ok {
		return unavailable(LogUnsupported), nil
	}

	token, err := s.repos.RepoToken(ctx, repo.ID)
	if err != nil {
		return JobLogResult{}, fmt.Errorf("get repo token: %w", err)
	}

	tail, err := reader.JobLogTail(ctx, JobLogRequest{
		InstanceURL: repo.ForgejoInstanceURL,
		Identifier:  repo.Identifier,
		Token:       token,
		ForgeJobID:  location.ForgeJobID,
		Lines:       lines,
	})

	switch {
	case err == nil:
		return JobLogResult{Available: true, Lines: tail.Lines, Truncated: tail.Truncated, ForgeURL: location.ForgeURL}, nil
	case errors.Is(err, ErrLogUnsupported):
		return unavailable(LogUnsupported), nil
	case errors.Is(err, ErrLogExpired):
		return unavailable(LogExpired), nil
	case errors.Is(err, ErrLogForbidden):
		return unavailable(LogForbidden), nil
	case errors.Is(err, ErrLogUnreachable):
		return unavailable(LogUnreachable), nil
	default:
		return JobLogResult{}, fmt.Errorf("read job log: %w", err)
	}
}

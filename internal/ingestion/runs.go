package ingestion

import (
	"context"
	"time"
)

// Run is one execution of a pipeline (a GitHub/Forgejo Actions workflow
// run).
type Run struct {
	ID           string
	RepoID       string
	ForgeRunID   string
	PipelineName string
	Status       string
	Conclusion   string
	StartedAt    *time.Time
	CompletedAt  *time.Time
	ForgeURL     string
}

// Job is one job within a Run.
type Job struct {
	ID          string
	RunID       string
	ForgeJobID  string
	Name        string
	Status      string
	Conclusion  string
	QueuedAt    *time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	ForgeURL    string
}

// Step is one step within a Job.
type Step struct {
	ID          string
	JobID       string
	Number      int
	Name        string
	Status      string
	Conclusion  string
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// RunStore is the port webhook ingestion persists runs, jobs, and steps
// through.
type RunStore interface {
	// RepoByIdentifier finds a tracked repo by forge and identifier
	// (owner/name) -- how an incoming webhook is resolved to the repo it
	// belongs to.
	RepoByIdentifier(ctx context.Context, forge Forge, identifier string) (Repo, error)
	// UpsertRun creates or updates a run, matched by (repo id, forge run
	// id), and returns the stored row (with its internal ID).
	UpsertRun(ctx context.Context, run Run) (Run, error)
	// UpsertJob creates or updates a job, matched by (run id, forge job
	// id), and returns the stored row (with its internal ID).
	UpsertJob(ctx context.Context, job Job) (Job, error)
	// ReplaceSteps replaces every step belonging to jobID with steps -- a
	// workflow_job event always carries the job's full, current step list,
	// so there's nothing to reconcile incrementally.
	ReplaceSteps(ctx context.Context, jobID string, steps []Step) error
}

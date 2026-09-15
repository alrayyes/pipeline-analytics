package ingestion_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"hello":"world"}`)
	secret := "shh"

	t.Run("accepts a valid signature with the sha256= prefix", func(t *testing.T) {
		t.Parallel()

		require.True(t, ingestion.VerifySignature(secret, body, sign(secret, body)))
	})

	t.Run("accepts a valid signature with no prefix (Forgejo's format)", func(t *testing.T) {
		t.Parallel()

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		require.True(t, ingestion.VerifySignature(secret, body, hex.EncodeToString(mac.Sum(nil))))
	})

	t.Run("rejects a signature from the wrong secret", func(t *testing.T) {
		t.Parallel()

		require.False(t, ingestion.VerifySignature(secret, body, sign("wrong-secret", body)))
	})

	t.Run("rejects a tampered body", func(t *testing.T) {
		t.Parallel()

		sig := sign(secret, body)
		require.False(t, ingestion.VerifySignature(secret, []byte(`{"hello":"tampered"}`), sig))
	})

	t.Run("rejects garbage", func(t *testing.T) {
		t.Parallel()

		require.False(t, ingestion.VerifySignature(secret, body, "not-hex-at-all!!"))
	})
}

type fakeRunStore struct {
	mu    sync.Mutex
	runs  map[string]ingestion.Run // key: repoID/forgeRunID
	jobs  map[string]ingestion.Job // key: runID/forgeJobID
	steps map[string][]ingestion.Step
	seq   int
}

func newFakeRunStore() *fakeRunStore {
	return &fakeRunStore{
		runs:  map[string]ingestion.Run{},
		jobs:  map[string]ingestion.Job{},
		steps: map[string][]ingestion.Step{},
	}
}

func (f *fakeRunStore) RepoByIdentifier(context.Context, ingestion.Forge, string) (ingestion.Repo, error) {
	return ingestion.Repo{}, errRepoNotFoundFake
}

func (f *fakeRunStore) UpsertRun(_ context.Context, run ingestion.Run) (ingestion.Run, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := run.RepoID + "/" + run.ForgeRunID
	if existing, ok := f.runs[key]; ok {
		run.ID = existing.ID
	} else {
		f.seq++
		run.ID = fmt.Sprintf("run-%d", f.seq)
	}

	f.runs[key] = run

	return run, nil
}

func (f *fakeRunStore) UpsertJob(_ context.Context, job ingestion.Job) (ingestion.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := job.RunID + "/" + job.ForgeJobID
	if existing, ok := f.jobs[key]; ok {
		job.ID = existing.ID
	} else {
		f.seq++
		job.ID = fmt.Sprintf("job-%d", f.seq)
	}

	f.jobs[key] = job

	return job, nil
}

func (f *fakeRunStore) ReplaceSteps(_ context.Context, jobID string, steps []ingestion.Step) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.steps[jobID] = steps

	return nil
}

var errRepoNotFoundFake = repoNotFoundFakeError{}

type repoNotFoundFakeError struct{}

func (repoNotFoundFakeError) Error() string { return "repo not found" }

const githubWorkflowRunPayload = `{
  "action": "completed",
  "workflow_run": {
    "id": 1001,
    "name": "CI",
    "status": "completed",
    "conclusion": "success",
    "run_started_at": "2026-09-15T10:00:00Z",
    "updated_at": "2026-09-15T10:05:00Z",
    "html_url": "https://github.com/alrayyes/pipeline-analytics/actions/runs/1001"
  },
  "repository": {"full_name": "alrayyes/pipeline-analytics"}
}`

const githubWorkflowJobPayload = `{
  "action": "completed",
  "workflow_job": {
    "id": 5001,
    "run_id": 1001,
    "workflow_name": "CI",
    "name": "build, vet, test",
    "status": "completed",
    "conclusion": "failure",
    "created_at": "2026-09-15T09:59:00Z",
    "started_at": "2026-09-15T10:00:00Z",
    "completed_at": "2026-09-15T10:04:00Z",
    "html_url": "https://github.com/alrayyes/pipeline-analytics/actions/runs/1001/job/5001",
    "steps": [
      {"name": "checkout", "status": "completed", "conclusion": "success", "number": 1, "started_at": "2026-09-15T10:00:00Z", "completed_at": "2026-09-15T10:00:10Z"},
      {"name": "go test", "status": "completed", "conclusion": "failure", "number": 2, "started_at": "2026-09-15T10:00:10Z", "completed_at": "2026-09-15T10:04:00Z"}
    ]
  },
  "repository": {"full_name": "alrayyes/pipeline-analytics"}
}`

func TestProcessGitHubEvent_WorkflowRun(t *testing.T) {
	t.Parallel()

	store := newFakeRunStore()
	repo := ingestion.Repo{ID: "repo-1", Forge: ingestion.ForgeGitHub}

	err := ingestion.ProcessGitHubEvent(context.Background(), store, repo, "workflow_run", []byte(githubWorkflowRunPayload))
	require.NoError(t, err)

	run, ok := store.runs["repo-1/1001"]
	require.True(t, ok)
	require.Equal(t, "CI", run.PipelineName)
	require.Equal(t, "completed", run.Status)
	require.Equal(t, "success", run.Conclusion)
	require.NotNil(t, run.StartedAt)
	require.NotNil(t, run.CompletedAt)
}

func TestProcessGitHubEvent_WorkflowJob(t *testing.T) {
	t.Parallel()

	store := newFakeRunStore()
	repo := ingestion.Repo{ID: "repo-1", Forge: ingestion.ForgeGitHub}

	t.Run("creates a run stub if the parent run hasn't arrived yet", func(t *testing.T) {
		err := ingestion.ProcessGitHubEvent(context.Background(), store, repo, "workflow_job", []byte(githubWorkflowJobPayload))
		require.NoError(t, err)

		run, ok := store.runs["repo-1/1001"]
		require.True(t, ok)
		require.Equal(t, "CI", run.PipelineName)

		job, ok := store.jobs[run.ID+"/5001"]
		require.True(t, ok)
		require.Equal(t, "build, vet, test", job.Name)
		require.Equal(t, "failure", job.Conclusion)

		steps := store.steps[job.ID]
		require.Len(t, steps, 2)
		require.Equal(t, "checkout", steps[0].Name)
		require.Equal(t, "failure", steps[1].Conclusion)
	})
}

func TestProcessGitHubEvent_UnknownEventType(t *testing.T) {
	t.Parallel()

	store := newFakeRunStore()
	repo := ingestion.Repo{ID: "repo-1"}

	err := ingestion.ProcessGitHubEvent(context.Background(), store, repo, "push", []byte(`{}`))
	require.NoError(t, err) // ignored, not an error
	require.Empty(t, store.runs)
}

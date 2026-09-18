package ingestion_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
	"github.com/stretchr/testify/require"
)

var errForgeUnavailable = errors.New("forge unavailable")

func TestReconciler_ReconcileRepo(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)

	t.Run("a first poll backfills runs, jobs, and steps and persists the new ETag", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		repo, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "alrayyes/pipeline-analytics"})
		require.NoError(t, err)

		runStore := newFakeRunStore()
		client := &fakeForgeClient{listRunsFunc: func(_ context.Context, req ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
			require.Empty(t, req.ETag) // repo's first poll -- no prior ETag to send

			return ingestion.ListRunsResult{
				ETag: `"v1"`,
				Runs: []ingestion.RunSnapshot{{
					ForgeRunID:   "1001",
					PipelineName: "CI",
					Status:       "completed",
					Conclusion:   "success",
					StartedAt:    &startedAt,
					Jobs: []ingestion.JobSnapshot{{
						ForgeJobID: "5001",
						Name:       "build",
						Status:     "completed",
						Conclusion: "success",
						Steps: []ingestion.StepSnapshot{
							{Number: 1, Name: "checkout", Status: "completed", Conclusion: "success"},
						},
					}},
				}},
			}, nil
		}}

		reconciler := ingestion.NewReconciler(store, runStore, map[ingestion.Forge]ingestion.ForgeClient{ingestion.ForgeGitHub: client})
		require.NoError(t, reconciler.ReconcileRepo(context.Background(), repo))

		run, ok := runStore.runs[repo.ID+"/1001"]
		require.True(t, ok)
		require.Equal(t, "CI", run.PipelineName)

		job, ok := runStore.jobs[run.ID+"/5001"]
		require.True(t, ok)
		require.Equal(t, "build", job.Name)
		require.Len(t, runStore.steps[job.ID], 1)

		updated, err := store.GetRepo(context.Background(), repo.ID)
		require.NoError(t, err)
		require.Equal(t, `"v1"`, updated.ReconcileETag)
	})

	t.Run("passes already-stored run states as KnownRuns on the next poll", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		repo, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "alrayyes/pipeline-analytics"})
		require.NoError(t, err)

		runStore := newFakeRunStore()
		_, err = runStore.UpsertRun(context.Background(), ingestion.Run{
			RepoID: repo.ID, ForgeRunID: "1001", PipelineName: "CI",
			Status: "completed", Conclusion: "success",
		})
		require.NoError(t, err)

		var gotKnownRuns map[string]ingestion.RunState

		client := &fakeForgeClient{listRunsFunc: func(_ context.Context, req ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
			gotKnownRuns = req.KnownRuns

			return ingestion.ListRunsResult{NotModified: true}, nil
		}}

		reconciler := ingestion.NewReconciler(store, runStore, map[ingestion.Forge]ingestion.ForgeClient{ingestion.ForgeGitHub: client})
		require.NoError(t, reconciler.ReconcileRepo(context.Background(), repo))

		require.Equal(t, map[string]ingestion.RunState{
			"1001": {Status: "completed", Conclusion: "success"},
		}, gotKnownRuns)
	})

	t.Run("an unchanged repo makes no run-storage writes", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		repo, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "alrayyes/pipeline-analytics"})
		require.NoError(t, err)
		require.NoError(t, store.SetReconcileETag(context.Background(), repo.ID, `"v1"`))
		repo, err = store.GetRepo(context.Background(), repo.ID)
		require.NoError(t, err)

		runStore := newFakeRunStore()
		client := &fakeForgeClient{listRunsFunc: func(_ context.Context, req ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
			require.Equal(t, `"v1"`, req.ETag)

			return ingestion.ListRunsResult{NotModified: true}, nil
		}}

		reconciler := ingestion.NewReconciler(store, runStore, map[ingestion.Forge]ingestion.ForgeClient{ingestion.ForgeGitHub: client})
		require.NoError(t, reconciler.ReconcileRepo(context.Background(), repo))

		require.Empty(t, runStore.runs)

		updated, err := store.GetRepo(context.Background(), repo.ID)
		require.NoError(t, err)
		require.Equal(t, `"v1"`, updated.ReconcileETag) // unchanged
	})

	t.Run("a repo whose forge has no registered client is skipped without error", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		repo, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeForgejo, Identifier: "alrayyes/pipeline-analytics"})
		require.NoError(t, err)

		reconciler := ingestion.NewReconciler(store, newFakeRunStore(), map[ingestion.Forge]ingestion.ForgeClient{})
		require.NoError(t, reconciler.ReconcileRepo(context.Background(), repo))
	})
}

func TestReconciler_Run(t *testing.T) {
	t.Parallel()

	t.Run("polls immediately and again on every tick, until the context is done", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		_, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "alrayyes/pipeline-analytics"})
		require.NoError(t, err)

		var pollCount atomic.Int64

		client := &fakeForgeClient{listRunsFunc: func(context.Context, ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
			pollCount.Add(1)

			return ingestion.ListRunsResult{NotModified: true}, nil
		}}

		reconciler := ingestion.NewReconciler(store, newFakeRunStore(), map[ingestion.Forge]ingestion.ForgeClient{ingestion.ForgeGitHub: client})

		ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
		defer cancel()

		reconciler.Run(ctx, 20*time.Millisecond)

		require.GreaterOrEqual(t, pollCount.Load(), int64(2)) // the immediate poll plus at least one tick
	})
}

func TestReconciler_ReconcileAll(t *testing.T) {
	t.Parallel()

	t.Run("one repo's failure doesn't block reconciling the others", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		failing, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "alrayyes/broken"})
		require.NoError(t, err)
		healthy, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "alrayyes/ok"})
		require.NoError(t, err)

		polled := map[string]bool{}
		client := &fakeForgeClient{listRunsFunc: func(_ context.Context, req ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
			polled[req.Identifier] = true
			if req.Identifier == failing.Identifier {
				return ingestion.ListRunsResult{}, errForgeUnavailable
			}

			return ingestion.ListRunsResult{NotModified: true}, nil
		}}

		reconciler := ingestion.NewReconciler(store, newFakeRunStore(), map[ingestion.Forge]ingestion.ForgeClient{ingestion.ForgeGitHub: client})
		reconciler.ReconcileAll(context.Background())

		require.True(t, polled[failing.Identifier])
		require.True(t, polled[healthy.Identifier])
	})
}

// TestReconciler_ReconcileRepo_SkipsUnchangedRunJobs wires a real
// github.Client (not the in-package fake) into Reconciler, since "zero jobs
// requests" is an HTTP-level guarantee only the real client's KnownRuns
// handling can demonstrate end to end -- skip-refetch-completed-run-jobs's
// whole point.
func TestReconciler_ReconcileRepo_SkipsUnchangedRunJobs(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	repo, err := store.CreateRepo(context.Background(), ingestion.NewRepo{Forge: ingestion.ForgeGitHub, Identifier: "alrayyes/pipeline-analytics"})
	require.NoError(t, err)

	runStore := newFakeRunStore()
	_, err = runStore.UpsertRun(context.Background(), ingestion.Run{
		RepoID: repo.ID, ForgeRunID: "1001", PipelineName: "CI",
		Status: "completed", Conclusion: "success",
	})
	require.NoError(t, err)

	jobsRequests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/repos/alrayyes/pipeline-analytics/actions/runs":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"workflow_runs": [
					{
						"id": 1001,
						"name": "CI",
						"status": "completed",
						"conclusion": "success",
						"run_started_at": "2026-09-15T10:00:00Z",
						"updated_at": "2026-09-15T10:05:00Z",
						"html_url": "https://github.com/alrayyes/pipeline-analytics/actions/runs/1001"
					}
				]
			}`))
		case "/repos/alrayyes/pipeline-analytics/actions/runs/1001/jobs":
			jobsRequests++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"jobs": []}`))
		default:
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := github.NewClient(server.URL + "/")
	require.NoError(t, err)

	reconciler := ingestion.NewReconciler(store, runStore, map[ingestion.Forge]ingestion.ForgeClient{ingestion.ForgeGitHub: client})
	require.NoError(t, reconciler.ReconcileRepo(context.Background(), repo))

	require.Zero(t, jobsRequests, "run 1001's status/conclusion already matched what was stored")
}

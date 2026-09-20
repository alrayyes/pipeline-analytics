package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

func at(minutesFromEpoch int) *time.Time {
	t := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(minutesFromEpoch) * time.Minute)

	return &t
}

// seedRepo registers alrayyes/pipeline-analytics through the API and
// returns its id.
func seedRepo(t *testing.T, srv testServer) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"forge":      "github",
		"identifier": "alrayyes/pipeline-analytics",
		"token":      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", strings.NewReader(string(body))))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var repo map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &repo))

	id, _ := repo["id"].(string)
	require.NotEmpty(t, id)

	return id
}

const seedRunPipelineName = "CI"

// pipelineListResponse mirrors the handler's wrapped GET /api/pipelines
// response shape, kept loose (map per pipeline) since these tests only
// assert on a handful of fields.
type pipelineListResponse struct {
	Pipelines []map[string]any `json:"pipelines"`
	HasMore   bool             `json:"hasMore"`
}

// seedOtherRepo registers a second, Forgejo-hosted repo through the API and
// returns its id -- for tests exercising the repoId/forge filters.
func seedOtherRepo(t *testing.T, srv testServer) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{ //nolint:gosec // test fixture value, not a real credential
		"forge":              "forgejo",
		"identifier":         "alrayyes/dotfiles",
		"forgejoInstanceUrl": "https://git.higherlearning.eu",
		"token":              "forgejo-token-5678",
	})
	require.NoError(t, err)

	req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", strings.NewReader(string(body))))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var repo map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &repo))

	id, _ := repo["id"].(string)
	require.NotEmpty(t, id)

	return id
}

// seedRun inserts a completed, successful run for repoID directly through
// the test server's underlying RunStore, mirroring how a real webhook
// delivery or reconciliation poll would populate the same tables.
func seedRun(t *testing.T, srv testServer, repoID, forgeRunID string, startedMin, durationMin int) ingestion.Run {
	t.Helper()

	run, err := srv.runStore.UpsertRun(context.Background(), ingestion.Run{
		RepoID:       repoID,
		ForgeRunID:   forgeRunID,
		PipelineName: seedRunPipelineName,
		Status:       "completed",
		Conclusion:   "success",
		StartedAt:    at(startedMin),
		CompletedAt:  at(startedMin + durationMin),
		ForgeURL:     "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + forgeRunID,
	})
	require.NoError(t, err)

	return run
}

func TestPipelinesList(t *testing.T) {
	t.Parallel()

	t.Run("reports every pipeline's health status", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRun(t, srv, repoID, "1", 0, 5)
		seedRun(t, srv, repoID, "2", 10, 5)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var got pipelineListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.False(t, got.HasMore)
		require.Len(t, got.Pipelines, 1)
		require.Equal(t, seedRunPipelineName, got.Pipelines[0]["name"])
		require.Equal(t, "healthy", got.Pipelines[0]["healthStatus"])
	})

	t.Run("reports the most recent run's start time as lastRunAt", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRun(t, srv, repoID, "1", 0, 5)
		seedRun(t, srv, repoID, "2", 10, 5)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var got pipelineListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got.Pipelines, 1)
		require.Equal(t, at(10).Format(time.RFC3339), got.Pipelines[0]["lastRunAt"])
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/pipelines", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("list paginates with limit and offset", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRun(t, srv, repoID, "1", 0, 5)

		otherRepoID := seedOtherRepo(t, srv)
		seedRun(t, srv, otherRepoID, "2", 10, 5)

		firstPage := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines?limit=1&offset=0", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, firstPage)

		require.Equal(t, http.StatusOK, rec.Code)

		var got pipelineListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.True(t, got.HasMore)
		require.Len(t, got.Pipelines, 1)

		secondPage := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines?limit=1&offset=1", nil))
		rec = httptest.NewRecorder()
		srv.ServeHTTP(rec, secondPage)

		require.Equal(t, http.StatusOK, rec.Code)

		got = pipelineListResponse{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.False(t, got.HasMore)
		require.Len(t, got.Pipelines, 1)
	})

	t.Run("list filters by the repoId query param", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRun(t, srv, repoID, "1", 0, 5)

		otherRepoID := seedOtherRepo(t, srv)
		seedRun(t, srv, otherRepoID, "2", 10, 5)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines?repoId="+repoID, nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var got pipelineListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got.Pipelines, 1)
		require.Equal(t, repoID, got.Pipelines[0]["repoId"])
	})

	t.Run("list filters by the forge query param", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRun(t, srv, repoID, "1", 0, 5)

		otherRepoID := seedOtherRepo(t, srv)
		seedRun(t, srv, otherRepoID, "2", 10, 5)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines?forge=forgejo", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var got pipelineListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got.Pipelines, 1)
		require.Equal(t, otherRepoID, got.Pipelines[0]["repoId"])
	})
}

func TestPipelineGet(t *testing.T) {
	t.Parallel()

	t.Run("returns pipeline detail with a duration trend", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRun(t, srv, repoID, "1", 0, 5)

		pipelineID := pipelineIDFromList(t, srv)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines/"+pipelineID, nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var detail map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detail))
		require.Equal(t, seedRunPipelineName, detail["name"])
		require.NotNil(t, detail["durationTrend"])
	})

	t.Run("an unknown pipeline is not found", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines/bm9wZQ", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPipelineSteps(t *testing.T) {
	t.Parallel()

	t.Run("ranks steps by duration contribution", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		run := seedRun(t, srv, repoID, "1", 0, 30)

		job, err := srv.runStore.UpsertJob(context.Background(), ingestion.Job{
			RunID:       run.ID,
			ForgeJobID:  "100",
			Name:        "build",
			Status:      "completed",
			Conclusion:  "success",
			QueuedAt:    at(0),
			StartedAt:   at(0),
			CompletedAt: at(30),
			ForgeURL:    "https://github.com/alrayyes/pipeline-analytics/actions/runs/1/job/100",
		})
		require.NoError(t, err)

		err = srv.runStore.ReplaceSteps(context.Background(), job.ID, []ingestion.Step{
			{Number: 1, Name: "checkout", Status: "completed", Conclusion: "success", StartedAt: at(0), CompletedAt: at(1)},
			{Number: 2, Name: "test", Status: "completed", Conclusion: "success", StartedAt: at(1), CompletedAt: at(20)},
		})
		require.NoError(t, err)

		pipelineID := pipelineIDFromList(t, srv)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines/"+pipelineID+"/steps", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var steps []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &steps))
		require.Len(t, steps, 2)
		require.Equal(t, "test", steps[0]["name"]) // longest-running step ranks first
	})
}

func TestFlakyRuns(t *testing.T) {
	t.Parallel()

	t.Run("lists only the runs where the named step failed, most-recent-first", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)

		passingRun := seedRun(t, srv, repoID, "1", 0, 5)
		seedJobStep(t, srv, passingRun.ID, "100", "test", "success")

		firstFailingRun := seedRun(t, srv, repoID, "2", 10, 5)
		seedJobStep(t, srv, firstFailingRun.ID, "200", "test", "failure")

		secondFailingRun := seedRun(t, srv, repoID, "3", 20, 5)
		seedJobStep(t, srv, secondFailingRun.ID, "300", "test", "failure")

		pipelineID := pipelineIDFromList(t, srv)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines/"+pipelineID+"/flaky-runs?step=test", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var runs []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &runs))
		require.Len(t, runs, 2)
		require.Equal(t, secondFailingRun.ID, runs[0]["runId"])
		require.Equal(t, firstFailingRun.ID, runs[1]["runId"])
		require.Contains(t, runs[0]["forgeUrl"], "/job/300")
	})

	t.Run("a step with no failures reports an empty list", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		run := seedRun(t, srv, repoID, "1", 0, 5)
		seedJobStep(t, srv, run.ID, "100", "test", "success")

		pipelineID := pipelineIDFromList(t, srv)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines/"+pipelineID+"/flaky-runs?step=test", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var runs []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &runs))
		require.Empty(t, runs)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/pipelines/bm9wZQ/flaky-runs?step=test", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestRunSteps(t *testing.T) {
	t.Parallel()

	t.Run("returns the run's own steps", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		run := seedRun(t, srv, repoID, "1", 0, 5)
		seedJobStep(t, srv, run.ID, "100", "checkout", "success")

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs/"+run.ID+"/steps", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var detail map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detail))
		require.Equal(t, run.ID, detail["runId"])

		steps, _ := detail["steps"].([]any)
		require.Len(t, steps, 1)
		step, _ := steps[0].(map[string]any)
		require.Equal(t, "checkout", step["name"])
		require.Contains(t, step["forgeUrl"], "/job/100")
	})

	t.Run("an unknown run is not found", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs/does-not-exist/steps", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/runs/does-not-exist/steps", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// TestPipelineHealth_SelfHeals pins down behavior the flaky-runs drill-down
// (#216) relies on already being true: health is recomputed live from the
// window on every request, nothing persisted, so a step's failure aging out
// of the window flips the pipeline back to healthy on its own.
func TestPipelineHealth_SelfHeals(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)

	failingRun := seedRun(t, srv, repoID, "1", 0, 5)
	seedJobStep(t, srv, failingRun.ID, "100", "test", "failure")

	passingRun := seedRun(t, srv, repoID, "2", 10, 5)
	seedJobStep(t, srv, passingRun.ID, "200", "test", "success")

	t.Run("the failure still counts within a window wide enough to include it", func(t *testing.T) {
		t.Parallel()

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines?window=2", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		var got pipelineListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, "unhealthy", got.Pipelines[0]["healthStatus"])
	})

	t.Run("once the failure ages out of a narrower window, the pipeline reports healthy again", func(t *testing.T) {
		t.Parallel()

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines?window=1", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		var got pipelineListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, "healthy", got.Pipelines[0]["healthStatus"])
	})
}

// seedJobStep records a completed job with a single step -- the shared
// shape TestFlakyRuns/TestRunSteps/TestPipelineHealth_SelfHeals need,
// where TestPipelineSteps and TestUnhealthySteps below construct jobs with
// more than one step by hand instead.
func seedJobStep(t *testing.T, srv testServer, runID, forgeJobID, stepName, conclusion string) {
	t.Helper()

	job, err := srv.runStore.UpsertJob(context.Background(), ingestion.Job{
		RunID:       runID,
		ForgeJobID:  forgeJobID,
		Name:        "build",
		Status:      "completed",
		Conclusion:  conclusion,
		QueuedAt:    at(0),
		StartedAt:   at(0),
		CompletedAt: at(5),
		ForgeURL:    "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + runID + "/job/" + forgeJobID,
	})
	require.NoError(t, err)

	err = srv.runStore.ReplaceSteps(context.Background(), job.ID, []ingestion.Step{
		{Number: 1, Name: stepName, Status: "completed", Conclusion: conclusion, StartedAt: at(0), CompletedAt: at(1)},
	})
	require.NoError(t, err)
}

func TestUnhealthySteps(t *testing.T) {
	t.Parallel()

	t.Run("reports only flaky or failing steps, grouped by pipeline", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		run := seedRun(t, srv, repoID, "1", 0, 30)

		job, err := srv.runStore.UpsertJob(context.Background(), ingestion.Job{
			RunID:       run.ID,
			ForgeJobID:  "100",
			Name:        "build",
			Status:      "completed",
			Conclusion:  "failure",
			QueuedAt:    at(0),
			StartedAt:   at(0),
			CompletedAt: at(30),
			ForgeURL:    "https://github.com/alrayyes/pipeline-analytics/actions/runs/1/job/100",
		})
		require.NoError(t, err)

		err = srv.runStore.ReplaceSteps(context.Background(), job.ID, []ingestion.Step{
			{Number: 1, Name: "checkout", Status: "completed", Conclusion: "success", StartedAt: at(0), CompletedAt: at(1)},
			{Number: 2, Name: "deploy", Status: "completed", Conclusion: "failure", StartedAt: at(1), CompletedAt: at(20)},
		})
		require.NoError(t, err)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/steps/unhealthy", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var groups []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &groups))
		require.Len(t, groups, 1)
		require.Equal(t, seedRunPipelineName, groups[0]["pipelineName"])
		require.Equal(t, repoID, groups[0]["repoId"])

		steps, _ := groups[0]["steps"].([]any)
		require.Len(t, steps, 1) // "checkout" never failed, so it's excluded
		step, _ := steps[0].(map[string]any)
		require.Equal(t, "deploy", step["name"])
	})

	t.Run("no unhealthy steps anywhere reports an empty list", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		run := seedRun(t, srv, repoID, "1", 0, 30)

		job, err := srv.runStore.UpsertJob(context.Background(), ingestion.Job{
			RunID:       run.ID,
			ForgeJobID:  "100",
			Name:        "build",
			Status:      "completed",
			Conclusion:  "success",
			QueuedAt:    at(0),
			StartedAt:   at(0),
			CompletedAt: at(30),
			ForgeURL:    "https://github.com/alrayyes/pipeline-analytics/actions/runs/1/job/100",
		})
		require.NoError(t, err)

		err = srv.runStore.ReplaceSteps(context.Background(), job.ID, []ingestion.Step{
			{Number: 1, Name: "checkout", Status: "completed", Conclusion: "success", StartedAt: at(0), CompletedAt: at(1)},
		})
		require.NoError(t, err)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/steps/unhealthy", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var groups []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &groups))
		require.Empty(t, groups)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/steps/unhealthy", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestRepoUsage(t *testing.T) {
	t.Parallel()

	t.Run("reports runner minutes broken down by workflow", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		run := seedRun(t, srv, repoID, "1", 0, 10)

		_, err := srv.runStore.UpsertJob(context.Background(), ingestion.Job{
			RunID:       run.ID,
			ForgeJobID:  "100",
			Name:        "build",
			Status:      "completed",
			Conclusion:  "success",
			StartedAt:   at(0),
			CompletedAt: at(10),
			ForgeURL:    "https://github.com/alrayyes/pipeline-analytics/actions/runs/1/job/100",
		})
		require.NoError(t, err)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/repos/"+repoID+"/usage", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var usage []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &usage))
		require.Len(t, usage, 1)
		require.Equal(t, seedRunPipelineName, usage[0]["workflow"])
		require.InDelta(t, 10.0, usage[0]["runnerMinutes"], 0.01)
	})

	t.Run("an unknown repo is not found", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/repos/does-not-exist/usage", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// pipelineIDFromList looks up the pipeline id GET /api/pipelines reports
// for the single pipeline a test seeded -- the id is an implementation
// detail the tests otherwise have no need to construct by hand.
func pipelineIDFromList(t *testing.T, srv testServer) string {
	t.Helper()

	req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines", nil))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var got pipelineListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.NotEmpty(t, got.Pipelines)

	id, _ := got.Pipelines[0]["id"].(string)
	require.NotEmpty(t, id)

	return id
}

package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

// seedRecentRun stores a concluded run that completed `endedAgo` before now,
// after running for ten minutes. The failure insights are windowed on real
// time, so these tests seed relative to the clock rather than the fixed
// epoch the other fixtures use.
func seedRecentRun(t *testing.T, srv testServer, repoID, pipeline, forgeRunID, conclusion string, endedAgo time.Duration) ingestion.Run {
	t.Helper()

	completed := time.Now().UTC().Add(-endedAgo)
	started := completed.Add(-10 * time.Minute)

	run, err := srv.runStore.UpsertRun(context.Background(), ingestion.Run{
		RepoID:       repoID,
		ForgeRunID:   forgeRunID,
		PipelineName: pipeline,
		Status:       "completed",
		Conclusion:   conclusion,
		StartedAt:    &started,
		CompletedAt:  &completed,
		ForgeURL:     "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + forgeRunID,
	})
	require.NoError(t, err)

	return run
}

func getFailureInsights(t *testing.T, srv testServer, query string) (int, map[string]any) {
	t.Helper()

	path := "/api/insights/failures"
	if query != "" {
		path += "?" + query
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, path, nil)))

	var body map[string]any
	if rec.Code == http.StatusOK {
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	}

	return rec.Code, body
}

func TestFailureInsights(t *testing.T) {
	t.Parallel()

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/insights/failures", nil))

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("reports run counts and the pass rate over the default 7d window", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRecentRun(t, srv, repoID, "CI", "1", "success", time.Hour)
		seedRecentRun(t, srv, repoID, "CI", "2", "success", 2*time.Hour)
		seedRecentRun(t, srv, repoID, "CI", "3", "success", 3*time.Hour)
		seedRecentRun(t, srv, repoID, "CI", "4", "failure", 4*time.Hour)

		code, body := getFailureInsights(t, srv, "")

		require.Equal(t, http.StatusOK, code)
		require.InDelta(t, 4, body["totalRuns"], 0)
		require.InDelta(t, 1, body["failedRuns"], 0)
		require.InDelta(t, 0.75, body["passRate"], 1e-9)
	})

	t.Run("the window parameter bounds which runs count", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRecentRun(t, srv, repoID, "CI", "1", "success", time.Hour)
		seedRecentRun(t, srv, repoID, "CI", "2", "failure", 30*time.Hour) // outside 24h, inside 7d

		_, day := getFailureInsights(t, srv, "window=24h")
		require.InDelta(t, 1, day["totalRuns"], 0)

		_, week := getFailureInsights(t, srv, "window=7d")
		require.InDelta(t, 2, week["totalRuns"], 0)
	})

	t.Run("an unrecognised window falls back to 7d, as the contract says", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRecentRun(t, srv, repoID, "CI", "1", "success", 30*time.Hour)

		code, body := getFailureInsights(t, srv, "window=banana")

		require.Equal(t, http.StatusOK, code)
		require.InDelta(t, 1, body["totalRuns"], 0)
	})

	t.Run("absent figures are omitted, not zero, and lists are empty arrays, not null", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		code, body := getFailureInsights(t, srv, "")

		require.Equal(t, http.StatusOK, code)
		require.NotContains(t, body, "passRate")
		require.NotContains(t, body, "passRateDelta")
		require.NotContains(t, body, "mttrSeconds")
		require.Equal(t, []any{}, body["stageDistribution"])
		require.Equal(t, []any{}, body["categoryBreakdown"])
		require.Equal(t, []any{}, body["topFailingPipelines"])
		require.Equal(t, []any{}, body["failureGroups"])
		require.InDelta(t, 0, body["flakyStepRatio"], 0)
	})

	t.Run("reports mean time to recovery in seconds", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRecentRun(t, srv, repoID, "CI", "1", "failure", 100*time.Minute)
		seedRecentRun(t, srv, repoID, "CI", "2", "success", 60*time.Minute)

		_, body := getFailureInsights(t, srv, "")

		require.InDelta(t, 40*60, body["mttrSeconds"], 1)
	})

	t.Run("ranks failing pipelines with ids the pipeline endpoints accept", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRecentRun(t, srv, repoID, "CI", "1", "failure", time.Hour)
		seedRecentRun(t, srv, repoID, "CI", "2", "failure", 2*time.Hour)
		seedRecentRun(t, srv, repoID, "Lint", "3", "failure", 3*time.Hour)
		seedRecentRun(t, srv, repoID, "Docs", "4", "success", 4*time.Hour)

		_, body := getFailureInsights(t, srv, "")

		top, _ := body["topFailingPipelines"].([]any)
		require.Len(t, top, 2)

		first, _ := top[0].(map[string]any)
		require.Equal(t, "CI", first["pipelineName"])
		require.Equal(t, repoID, first["repoId"])
		require.InDelta(t, 2, first["failedRuns"], 0)
		require.InDelta(t, 2, first["runs"], 0)

		id, _ := first["pipelineId"].(string)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/pipelines/"+id, nil)))
		require.Equal(t, http.StatusOK, rec.Code, "the id is the same one GET /api/pipelines/{id} takes")
	})

	t.Run("groups failed steps with a category, the stage distribution and the flaky ratio", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		failed := seedRecentRun(t, srv, repoID, "CI", "1", "failure", time.Hour)
		passed := seedRecentRun(t, srv, repoID, "CI", "2", "success", 2*time.Hour)
		seedJobStep(t, srv, failed.ID, "10", "Run unit tests", "failure")
		seedJobStep(t, srv, passed.ID, "11", "Run unit tests", "success")

		_, body := getFailureInsights(t, srv, "")

		groups, _ := body["failureGroups"].([]any)
		require.Len(t, groups, 1)

		group, _ := groups[0].(map[string]any)
		require.Equal(t, "Run unit tests", group["step"])
		require.Equal(t, "code_tests", group["category"])
		require.Equal(t, "failure", group["conclusion"])
		require.InDelta(t, 1, group["occurrences"], 0)

		pipelines, _ := group["pipelines"].([]any)
		require.Len(t, pipelines, 1)

		pipeline, _ := pipelines[0].(map[string]any)
		require.Equal(t, "CI", pipeline["pipelineName"])
		require.NotEmpty(t, pipeline["pipelineId"])

		require.Equal(t, []any{map[string]any{"step": "Run unit tests", "failures": float64(1), "share": float64(1)}}, body["stageDistribution"])
		require.Equal(t, []any{map[string]any{"category": "code_tests", "occurrences": float64(1), "share": float64(1)}}, body["categoryBreakdown"])
		require.InDelta(t, 1, body["flakyStepRatio"], 0, "it both passed and failed: the only step is flaky")
	})

	t.Run("scopes to one repo", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		otherID := seedOtherRepo(t, srv)
		seedRecentRun(t, srv, repoID, "CI", "1", "success", time.Hour)
		seedRecentRun(t, srv, otherID, "CI", "2", "failure", time.Hour)

		_, body := getFailureInsights(t, srv, "repoId="+repoID)

		require.InDelta(t, 1, body["totalRuns"], 0)
		require.InDelta(t, 0, body["failedRuns"], 0)
	})

	t.Run("scopes to one forge", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		seedRecentRun(t, srv, seedRepo(t, srv), "CI", "1", "success", time.Hour)
		seedRecentRun(t, srv, seedOtherRepo(t, srv), "CI", "2", "failure", time.Hour)

		_, github := getFailureInsights(t, srv, "forge=github")
		_, forgejo := getFailureInsights(t, srv, "forge=forgejo")

		require.InDelta(t, 0, github["failedRuns"], 0)
		require.InDelta(t, 1, forgejo["failedRuns"], 0)
		require.InDelta(t, 1, forgejo["totalRuns"], 0)
	})
}

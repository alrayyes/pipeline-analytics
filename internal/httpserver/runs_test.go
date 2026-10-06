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

type runListBody struct {
	Runs    []map[string]any `json:"runs"`
	HasMore bool             `json:"hasMore"`
}

func getRuns(t *testing.T, srv testServer, query string) (int, runListBody) {
	t.Helper()

	path := "/api/runs"
	if query != "" {
		path += "?" + query
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, path, nil)))

	var body runListBody
	if rec.Code == http.StatusOK {
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	}

	return rec.Code, body
}

// seedRunWith stores a run with the given state and commit fields, started
// startedMin minutes after the fixtures' fixed epoch.
func seedRunWith(t *testing.T, srv testServer, repoID string, run ingestion.Run) ingestion.Run {
	t.Helper()

	run.RepoID = repoID
	if run.PipelineName == "" {
		run.PipelineName = "CI"
	}

	if run.ForgeURL == "" {
		run.ForgeURL = "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + run.ForgeRunID
	}

	created, err := srv.runStore.UpsertRun(context.Background(), run)
	require.NoError(t, err)

	return created
}

func TestListRuns(t *testing.T) {
	t.Parallel()

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/runs", nil))

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("lists runs newest first with their commit fields, duration and forge link", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "success", StartedAt: at(10), CompletedAt: at(15)})
		seedRunWith(t, srv, repoID, ingestion.Run{
			ForgeRunID: "2", Status: "completed", Conclusion: "failure", StartedAt: at(20), CompletedAt: at(27),
			Branch: "main", SHA: "c4d291a", Message: "fix(stripe): webhook retry", Actor: "marcus-v",
		})

		code, body := getRuns(t, srv, "")

		require.Equal(t, http.StatusOK, code)
		require.Len(t, body.Runs, 2)

		newest := body.Runs[0]
		require.Equal(t, "CI", newest["pipelineName"])
		require.Equal(t, "failure", newest["conclusion"])
		require.Equal(t, "main", newest["branch"])
		require.Equal(t, "c4d291a", newest["sha"])
		require.Equal(t, "fix(stripe): webhook retry", newest["message"])
		require.Equal(t, "marcus-v", newest["actor"])
		require.Contains(t, newest["forgeUrl"], "/actions/runs/2")
		require.InDelta(t, 7*60, newest["durationSeconds"], 0)
		require.Equal(t, repoID, newest["repoId"])
	})

	t.Run("omits what isn't known instead of sending empty strings or zero", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "1", Status: "in_progress", StartedAt: at(10)})

		_, body := getRuns(t, srv, "")

		running := body.Runs[0]
		require.NotContains(t, running, "conclusion")
		require.NotContains(t, running, "durationSeconds")
		require.NotContains(t, running, "branch")
		require.NotContains(t, running, "sha")
		require.NotContains(t, running, "actor")
	})

	t.Run("a run with no steps has an empty array, not null", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		seedRunWith(t, srv, seedRepo(t, srv), ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "success", StartedAt: at(10), CompletedAt: at(11)})

		_, body := getRuns(t, srv, "")

		require.Equal(t, []any{}, body.Runs[0]["steps"])
	})

	t.Run("an empty database is an empty list, not null", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs", nil)))

		require.JSONEq(t, `{"runs":[],"hasMore":false}`, rec.Body.String())
	})

	t.Run("each run carries its steps in order, and its id works with the run-steps endpoint", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		run := seedRunWith(t, srv, seedRepo(t, srv), ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "failure", StartedAt: at(10), CompletedAt: at(15)})
		seedJobStep(t, srv, run.ID, "10", "build", "success")
		seedJobStep(t, srv, run.ID, "11", "canary", "failure")

		_, body := getRuns(t, srv, "")

		steps, _ := body.Runs[0]["steps"].([]any)
		require.Len(t, steps, 2)

		canary, _ := steps[1].(map[string]any)
		require.Equal(t, "canary", canary["name"])
		require.Equal(t, "failure", canary["conclusion"])

		id, _ := body.Runs[0]["id"].(string)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs/"+id+"/steps", nil)))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("every run and step carries a normalized outcome, and the run-steps endpoint does too", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		failed := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "failure", StartedAt: at(30), CompletedAt: at(31)})
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "2", Status: "in_progress", StartedAt: at(20)})
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "3", Status: "queued", StartedAt: at(10)})
		seedJobStep(t, srv, failed.ID, "10", "build", "success")
		seedJobStep(t, srv, failed.ID, "11", "deploy", "timed_out")

		_, body := getRuns(t, srv, "")

		outcomes := map[string]any{}
		for _, run := range body.Runs {
			outcomes[run["id"].(string)] = run["outcome"]
		}

		require.Equal(t, "failed", outcomes[failed.ID])
		require.ElementsMatch(t, []any{"failed", "running", "queued"}, []any{body.Runs[0]["outcome"], body.Runs[1]["outcome"], body.Runs[2]["outcome"]})

		var steps []any

		for _, run := range body.Runs {
			if run["id"] == failed.ID {
				steps, _ = run["steps"].([]any)
			}
		}

		require.Len(t, steps, 2)
		require.Equal(t, "passed", steps[0].(map[string]any)["outcome"])
		require.Equal(t, "failed", steps[1].(map[string]any)["outcome"], "a timed_out step is failed; its conclusion stays timed_out")
		require.Equal(t, "timed_out", steps[1].(map[string]any)["conclusion"])

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs/"+failed.ID+"/steps", nil)))
		require.Equal(t, http.StatusOK, rec.Code)

		var detail struct {
			Steps []map[string]any `json:"steps"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detail))
		require.Equal(t, "failed", detail.Steps[1]["outcome"])
	})

	t.Run("filters by status bucket", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "success", StartedAt: at(10), CompletedAt: at(11)})
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "2", Status: "completed", Conclusion: "failure", StartedAt: at(20), CompletedAt: at(21)})
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "3", Status: "in_progress", StartedAt: at(30)})

		for status, want := range map[string]string{"failed": "failure", "success": "success"} {
			_, body := getRuns(t, srv, "status="+status)
			require.Len(t, body.Runs, 1, status)
			require.Equal(t, want, body.Runs[0]["conclusion"], status)
		}

		_, running := getRuns(t, srv, "status=running")
		require.Len(t, running.Runs, 1)
		require.Equal(t, "in_progress", running.Runs[0]["status"])

		_, all := getRuns(t, srv, "status=all")
		require.Len(t, all.Runs, 3)
	})

	t.Run("an unknown status is a 400, not an empty list", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs?status=passed", nil)))

		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), `"code":"invalid_status"`)
	})

	t.Run("paginates and reports whether more remain", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)

		for i, id := range []string{"1", "2", "3"} {
			seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: id, Status: "completed", Conclusion: "success", StartedAt: at(10 * (i + 1)), CompletedAt: at(10*(i+1) + 1)})
		}

		_, first := getRuns(t, srv, "limit=2")
		require.Len(t, first.Runs, 2)
		require.True(t, first.HasMore)

		_, last := getRuns(t, srv, "limit=2&offset=2")
		require.Len(t, last.Runs, 1)
		require.False(t, last.HasMore)
	})

	t.Run("scopes to one repo and one forge", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		otherID := seedOtherRepo(t, srv)
		seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "success", StartedAt: at(10), CompletedAt: at(11)})
		seedRunWith(t, srv, otherID, ingestion.Run{ForgeRunID: "2", Status: "completed", Conclusion: "failure", StartedAt: at(20), CompletedAt: at(21)})

		_, byRepo := getRuns(t, srv, "repoId="+otherID)
		require.Len(t, byRepo.Runs, 1)
		require.Equal(t, otherID, byRepo.Runs[0]["repoId"])

		_, byForge := getRuns(t, srv, "forge=github")
		require.Len(t, byForge.Runs, 1)
		require.Equal(t, repoID, byForge.Runs[0]["repoId"])
	})

	t.Run("startedAt is an RFC 3339 time", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		seedRunWith(t, srv, seedRepo(t, srv), ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "success", StartedAt: at(10), CompletedAt: at(11)})

		_, body := getRuns(t, srv, "")

		startedAt, _ := body.Runs[0]["startedAt"].(string)
		parsed, err := time.Parse(time.RFC3339, startedAt)
		require.NoError(t, err)
		require.True(t, parsed.Equal(*at(10)))
	})
}

func TestListRunsActions(t *testing.T) {
	t.Parallel()

	actionsOf := func(t *testing.T, forgeRepo func(testServer) string, run ingestion.Run) any {
		t.Helper()

		srv := newTestServer(t, nil)
		seedRunWith(t, srv, forgeRepo(srv), run)

		_, body := getRuns(t, srv, "")
		require.Len(t, body.Runs, 1)

		return body.Runs[0]["actions"]
	}

	github := func(srv testServer) string { return seedRepo(t, srv) }
	forgejo := func(srv testServer) string { return seedOtherRepo(t, srv) }

	t.Run("a concluded GitHub run offers a re-run", func(t *testing.T) {
		t.Parallel()

		got := actionsOf(t, github, ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "failure", StartedAt: at(1), CompletedAt: at(2)})

		require.Equal(t, []any{"rerun"}, got)
	})

	t.Run("a running GitHub run offers a cancel", func(t *testing.T) {
		t.Parallel()

		got := actionsOf(t, github, ingestion.Run{ForgeRunID: "2", Status: "in_progress", StartedAt: at(1)})

		require.Equal(t, []any{"cancel"}, got)
	})

	t.Run("a queued GitHub run offers a cancel", func(t *testing.T) {
		t.Parallel()

		got := actionsOf(t, github, ingestion.Run{ForgeRunID: "3", Status: "queued", StartedAt: at(1)})

		require.Equal(t, []any{"cancel"}, got)
	})

	t.Run("a Forgejo run offers nothing, as an empty list", func(t *testing.T) {
		t.Parallel()

		got := actionsOf(t, forgejo, ingestion.Run{ForgeRunID: "4", Status: "completed", Conclusion: "failure", StartedAt: at(1), CompletedAt: at(2)})

		require.Equal(t, []any{}, got)
	})
}

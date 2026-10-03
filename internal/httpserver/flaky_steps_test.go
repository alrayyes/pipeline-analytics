package httpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type flakyStepsBody struct {
	Window string `json:"window"`
	Steps  []struct {
		PipelineID     string   `json:"pipelineId"`
		PipelineName   string   `json:"pipelineName"`
		RepoID         string   `json:"repoId"`
		Name           string   `json:"name"`
		FlakeRate      float64  `json:"flakeRate"`
		RunCount       int      `json:"runCount"`
		RecentOutcomes []string `json:"recentOutcomes"`
	} `json:"steps"`
	HasMore bool `json:"hasMore"`
}

func getFlakySteps(t *testing.T, srv testServer, query string) (int, flakyStepsBody) {
	t.Helper()

	path := "/api/steps/flaky"
	if query != "" {
		path += "?" + query
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, path, nil)))

	var body flakyStepsBody
	if rec.Code == http.StatusOK {
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	}

	return rec.Code, body
}

// seedFlakyHistory stores one run per conclusion, oldest first, each with a
// "test" step that ended the same way.
func seedFlakyHistory(t *testing.T, srv testServer, repoID string, conclusions ...string) {
	t.Helper()

	for i, conclusion := range conclusions {
		ago := time.Duration(len(conclusions)-i) * time.Hour
		id := string(rune('a' + i))
		run := seedRecentRun(t, srv, repoID, "CI", id, conclusion, ago)
		seedJobStep(t, srv, run.ID, id+"-job", "test", conclusion)
	}
}

func TestFlakySteps(t *testing.T) {
	t.Parallel()

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/steps/flaky", nil))

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("lists a flaky step with its rate, run count and outcomes, oldest first", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedFlakyHistory(t, srv, repoID, "success", "failure", "success", "success")

		code, body := getFlakySteps(t, srv, "")

		require.Equal(t, http.StatusOK, code)
		require.Equal(t, "7d", body.Window)
		require.Len(t, body.Steps, 1)
		require.Equal(t, "test", body.Steps[0].Name)
		require.Equal(t, "CI", body.Steps[0].PipelineName)
		require.Equal(t, repoID, body.Steps[0].RepoID)
		require.NotEmpty(t, body.Steps[0].PipelineID)
		require.InDelta(t, 0.25, body.Steps[0].FlakeRate, 1e-9)
		require.Equal(t, 4, body.Steps[0].RunCount)
		require.Equal(t, []string{"passed", "failed", "passed", "passed"}, body.Steps[0].RecentOutcomes)
		require.False(t, body.HasMore)
	})

	t.Run("a step that only ever fails or only ever passes isn't listed", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		seedFlakyHistory(t, srv, seedRepo(t, srv), "failure", "failure")

		_, body := getFlakySteps(t, srv, "")

		require.Empty(t, body.Steps)
	})

	t.Run("an empty result is an empty array, not null", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/steps/flaky", nil)))

		require.Contains(t, rec.Body.String(), `"steps":[]`)
	})

	t.Run("the window bounds which runs count, and an unknown one reports the default", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		seedFlakyHistory(t, srv, seedRepo(t, srv), "success", "failure") // 2h and 1h ago

		_, day := getFlakySteps(t, srv, "window=24h")
		require.Equal(t, "24h", day.Window)
		require.Len(t, day.Steps, 1)

		_, odd := getFlakySteps(t, srv, "window=bogus")
		require.Equal(t, "7d", odd.Window)
	})

	t.Run("pages the list", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)

		for _, name := range []string{"one", "two"} {
			ok := seedRecentRun(t, srv, repoID, name, "ok"+name, "success", 2*time.Hour)
			bad := seedRecentRun(t, srv, repoID, name, "bad"+name, "failure", time.Hour)
			seedJobStep(t, srv, ok.ID, "ok"+name, "test", "success")
			seedJobStep(t, srv, bad.ID, "bad"+name, "test", "failure")
		}

		_, first := getFlakySteps(t, srv, "limit=1")
		require.Len(t, first.Steps, 1)
		require.True(t, first.HasMore)

		_, second := getFlakySteps(t, srv, "limit=1&offset=1")
		require.Len(t, second.Steps, 1)
		require.False(t, second.HasMore)
		require.NotEqual(t, first.Steps[0].PipelineName, second.Steps[0].PipelineName)
	})
}

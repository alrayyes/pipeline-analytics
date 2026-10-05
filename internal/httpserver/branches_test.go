package httpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type branchListBody struct {
	Window   string `json:"window"`
	Branches []struct {
		Name     string `json:"name"`
		RunCount int    `json:"runCount"`
	} `json:"branches"`
}

func getBranches(t *testing.T, srv testServer, query string) (int, branchListBody) {
	t.Helper()

	path := "/api/branches"
	if query != "" {
		path += "?" + query
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, path, nil)))

	var body branchListBody
	if rec.Code == http.StatusOK {
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	}

	return rec.Code, body
}

// The branch selector (#344) offers the branches that have runs in the
// window, busiest first.
func TestBranches(t *testing.T) {
	t.Parallel()

	t.Run("lists branches with their run counts, busiest first then by name", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedBranchRun(t, srv, repoID, "1", "main", "success", time.Hour)
		seedBranchRun(t, srv, repoID, "2", "main", "failure", 2*time.Hour)
		seedBranchRun(t, srv, repoID, "3", "feature/b", "success", 3*time.Hour)
		seedBranchRun(t, srv, repoID, "4", "feature/a", "success", 4*time.Hour)

		code, body := getBranches(t, srv, "")
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, "7d", body.Window)
		require.Len(t, body.Branches, 3)
		require.Equal(t, "main", body.Branches[0].Name)
		require.Equal(t, 2, body.Branches[0].RunCount)
		require.Equal(t, "feature/a", body.Branches[1].Name)
		require.Equal(t, "feature/b", body.Branches[2].Name)
	})

	t.Run("a run with no recorded branch isn't a branch", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedBranchRun(t, srv, repoID, "1", "", "success", time.Hour)
		seedBranchRun(t, srv, repoID, "2", "main", "success", 2*time.Hour)

		_, body := getBranches(t, srv, "")
		require.Len(t, body.Branches, 1)
		require.Equal(t, "main", body.Branches[0].Name)
	})

	t.Run("the window bounds which runs count, and is reported", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedBranchRun(t, srv, repoID, "1", "main", "success", time.Hour)
		seedBranchRun(t, srv, repoID, "2", "old", "success", 3*24*time.Hour)

		_, day := getBranches(t, srv, "window=24h")
		require.Equal(t, "24h", day.Window)
		require.Len(t, day.Branches, 1)

		_, week := getBranches(t, srv, "window=7d")
		require.Len(t, week.Branches, 2)
	})

	t.Run("an unrecognised window falls back to 7d", func(t *testing.T) {
		t.Parallel()

		_, body := getBranches(t, newTestServer(t, nil), "window=bogus")
		require.Equal(t, "7d", body.Window)
	})

	t.Run("no runs is an empty list, not an error", func(t *testing.T) {
		t.Parallel()

		code, body := getBranches(t, newTestServer(t, nil), "")
		require.Equal(t, http.StatusOK, code)
		require.Empty(t, body.Branches)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		newTestServer(t, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/branches", nil))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestMCPListBranches(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	seedBranchRun(t, srv, repoID, "1", "main", "success", time.Hour)
	seedBranchRun(t, srv, repoID, "2", "main", "success", 2*time.Hour)
	seedBranchRun(t, srv, repoID, "3", "dev", "success", 3*time.Hour)

	got := callTool[branchListBody](t, connectMCP(t, srv), "list_branches", map[string]any{"window": "24h"})

	require.Equal(t, "24h", got.Window)
	require.Len(t, got.Branches, 2)
	require.Equal(t, "main", got.Branches[0].Name)
	require.Equal(t, 2, got.Branches[0].RunCount)
}

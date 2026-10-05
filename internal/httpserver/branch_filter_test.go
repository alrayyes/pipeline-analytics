package httpserver_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

// seedBranchRun stores a concluded run on branch that ended endedAgo before
// now, so the windowed insights see it.
func seedBranchRun(t *testing.T, srv testServer, repoID, forgeRunID, branch, conclusion string, endedAgo time.Duration) ingestion.Run {
	t.Helper()

	completed := time.Now().UTC().Add(-endedAgo)
	started := completed.Add(-10 * time.Minute)

	run, err := srv.runStore.UpsertRun(context.Background(), ingestion.Run{
		RepoID:       repoID,
		ForgeRunID:   forgeRunID,
		PipelineName: "CI",
		Status:       "completed",
		Conclusion:   conclusion,
		StartedAt:    &started,
		CompletedAt:  &completed,
		Branch:       branch,
		ForgeURL:     "https://github.com/alrayyes/pipeline-analytics/actions/runs/" + forgeRunID,
	})
	require.NoError(t, err)

	return run
}

// A broken feature branch shouldn't hide a healthy main (#344), so each
// telemetry read takes `branch` and covers only that branch's runs.
func TestBranchFilter(t *testing.T) {
	t.Parallel()

	t.Run("the failure insights cover only the chosen branch", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedBranchRun(t, srv, repoID, "1", "main", "success", time.Hour)
		seedBranchRun(t, srv, repoID, "2", "main", "success", 2*time.Hour)
		seedBranchRun(t, srv, repoID, "3", "feature/x", "failure", 3*time.Hour)
		seedBranchRun(t, srv, repoID, "4", "feature/x", "failure", 4*time.Hour)

		_, all := getFailureInsights(t, srv, "")
		require.InDelta(t, 4, all["totalRuns"], 0, "no selection includes every branch, as before")

		_, main := getFailureInsights(t, srv, "branch=main")
		require.InDelta(t, 2, main["totalRuns"], 0)
		require.InDelta(t, 0, main["failedRuns"], 0)

		_, feature := getFailureInsights(t, srv, "branch=feature%2Fx")
		require.InDelta(t, 2, feature["failedRuns"], 0)
	})

	t.Run("a branch with no runs is an empty result, not an error", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedBranchRun(t, srv, repoID, "1", "main", "success", time.Hour)

		code, body := getFailureInsights(t, srv, "branch=nope")
		require.Equal(t, http.StatusOK, code)
		require.InDelta(t, 0, body["totalRuns"], 0)

		runsCode, runs := getRuns(t, srv, "branch=nope")
		require.Equal(t, http.StatusOK, runsCode)
		require.Empty(t, runs.Runs)
	})

	t.Run("the run list covers only the chosen branch", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		onMain := seedBranchRun(t, srv, repoID, "1", "main", "success", time.Hour)
		seedBranchRun(t, srv, repoID, "2", "feature/x", "failure", 2*time.Hour)

		_, all := getRuns(t, srv, "")
		require.Len(t, all.Runs, 2)

		_, main := getRuns(t, srv, "branch=main")
		require.Len(t, main.Runs, 1)
		require.Equal(t, onMain.ID, main.Runs[0]["id"])
	})

	t.Run("flaky steps cover only the chosen branch", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)

		pass := seedBranchRun(t, srv, repoID, "1", "main", "success", time.Hour)
		fail := seedBranchRun(t, srv, repoID, "2", "main", "failure", 2*time.Hour)
		seedJobStep(t, srv, pass.ID, "10", "go test", "success")
		seedJobStep(t, srv, fail.ID, "11", "go test", "failure")

		other := seedBranchRun(t, srv, repoID, "3", "feature/x", "success", 3*time.Hour)
		seedJobStep(t, srv, other.ID, "12", "go test", "success")

		_, main := getFlakySteps(t, srv, "branch=main")
		require.Len(t, main.Steps, 1)
		require.Equal(t, 2, main.Steps[0].RunCount, "only main's two runs of the step")

		_, feature := getFlakySteps(t, srv, "branch=feature%2Fx")
		require.Empty(t, feature.Steps, "the step never failed on that branch, so it isn't flaky there")
	})

	t.Run("a branch longer than 255 characters is rejected, on every read", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		long := "branch=" + strings.Repeat("x", 256)

		code, _ := getFailureInsights(t, srv, long)
		require.Equal(t, http.StatusBadRequest, code)

		code, _ = getRuns(t, srv, long)
		require.Equal(t, http.StatusBadRequest, code)

		code, _ = getFlakySteps(t, srv, long)
		require.Equal(t, http.StatusBadRequest, code)
	})
}

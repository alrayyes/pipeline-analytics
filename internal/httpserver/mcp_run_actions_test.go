package httpserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// connectMCPWithBearer is connectMCP for a caller that holds an API token and
// no session, which is how an agent outside the dashboard reaches the server.
func connectMCPWithBearer(t *testing.T, srv testServer, token string) *mcp.ClientSession {
	t.Helper()

	httpSrv := httptest.NewServer(srv)
	t.Cleanup(httpSrv.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             httpSrv.URL + "/api/mcp",
		HTTPClient:           &http.Client{Transport: authRoundTripper{bearer: token}},
		DisableStandaloneSSE: true,
	}

	session, err := client.Connect(context.Background(), transport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	return session
}

// runActionError calls a run-action tool and returns the error it reported.
func runActionError(t *testing.T, session *mcp.ClientSession, tool, runID string) string {
	t.Helper()

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: map[string]any{"runId": runID}})
	require.NoError(t, err)
	require.True(t, res.IsError, "the call should have reported an error")

	return toolErrorText(t, res)
}

func TestMCPRerunRun(t *testing.T) {
	t.Parallel()

	t.Run("re-runs the failed jobs of a failed run through the same service", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "77", Status: "completed", Conclusion: "failure"})
		session := connectMCP(t, srv)

		got := callTool[map[string]any](t, session, "rerun_run", map[string]any{"runId": run.ID})

		require.Equal(t, true, got["accepted"])
		require.Equal(t, "rerun", actor.do)
		require.True(t, actor.got.FailedOnly)
		require.Equal(t, "77", actor.got.ForgeRunID)
	})

	t.Run("a run that has not concluded is not actionable, as on the endpoint", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "79", Status: "in_progress"})

		msg := runActionError(t, connectMCP(t, srv), "rerun_run", run.ID)

		require.Contains(t, msg, "not_actionable")
		require.Empty(t, actor.do, "the forge isn't asked")
	})

	t.Run("a token without Actions write is forbidden", func(t *testing.T) {
		t.Parallel()

		srv, repoID := actionTestServer(t, &stubRunActor{err: ingestion.ErrActionForbidden})
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "80", Status: "completed", Conclusion: "failure"})

		require.Contains(t, runActionError(t, connectMCP(t, srv), "rerun_run", run.ID), "forbidden")
	})

	t.Run("an unknown run is not found", func(t *testing.T) {
		t.Parallel()

		srv, _ := actionTestServer(t, &stubRunActor{})

		require.Contains(t, runActionError(t, connectMCP(t, srv), "rerun_run", "no-such-run"), "not_found")
	})

	t.Run("an API token is not a session, so it is refused before the forge is asked", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, authStore, repoID := actionTestServerWithAuthStore(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "81", Status: "completed", Conclusion: "failure"})

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID, 0)
		require.NoError(t, err)

		msg := runActionError(t, connectMCPWithBearer(t, srv, raw), "rerun_run", run.ID)

		require.Contains(t, msg, "unauthenticated")
		require.Empty(t, actor.do, "the forge isn't asked")
	})
}

func TestMCPCancelRun(t *testing.T) {
	t.Parallel()

	t.Run("cancels a run that has not concluded", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "90", Status: "in_progress"})

		got := callTool[map[string]any](t, connectMCP(t, srv), "cancel_run", map[string]any{"runId": run.ID})

		require.Equal(t, true, got["accepted"])
		require.Equal(t, "cancel", actor.do)
	})

	t.Run("a concluded run is not actionable", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "91", Status: "completed", Conclusion: "success"})

		require.Contains(t, runActionError(t, connectMCP(t, srv), "cancel_run", run.ID), "not_actionable")
		require.Empty(t, actor.do)
	})

	t.Run("a forge with no actor is unsupported", func(t *testing.T) {
		t.Parallel()

		srv, repoID := actionTestServer(t, nil)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "92", Status: "in_progress"})

		require.Contains(t, runActionError(t, connectMCP(t, srv), "cancel_run", run.ID), "unsupported")
	})
}

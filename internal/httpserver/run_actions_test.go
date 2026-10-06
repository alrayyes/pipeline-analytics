package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/stretchr/testify/require"
)

type stubRunActor struct {
	err error
	got ingestion.RunActionRequest
	do  string
}

func (s *stubRunActor) RerunRun(_ context.Context, req ingestion.RunActionRequest) error {
	s.got, s.do = req, "rerun"

	return s.err
}

func (s *stubRunActor) CancelRun(_ context.Context, req ingestion.RunActionRequest) error {
	s.got, s.do = req, "cancel"

	return s.err
}

// actionTestServer seeds one repo and returns the server and the repo's id.
// A nil actor leaves GitHub with no actor, as Forgejo has none.
func actionTestServer(t *testing.T, actor ingestion.RunActor) (testServer, string) {
	t.Helper()

	srv, _, repoID := actionTestServerWithAuthStore(t, actor)

	return srv, repoID
}

func actionTestServerWithAuthStore(t *testing.T, actor ingestion.RunActor) (testServer, auth.Store, string) {
	t.Helper()

	var authStore auth.Store

	srv := newTestServerWithDeps(t, func(deps *httpserver.Deps) {
		authStore = deps.AuthStore

		store, ok := deps.RunStore.(*ingestionsqlite.Store)
		require.True(t, ok)

		actors := map[ingestion.Forge]ingestion.RunActor{}
		if actor != nil {
			actors[ingestion.ForgeGitHub] = actor
		}

		deps.RunActions = ingestion.NewRunActionService(store, store, actors)
	})

	return srv, authStore, seedRepo(t, srv)
}

func postAction(srv testServer, runID, action string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/"+action, nil)))

	return rec
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Code string `json:"code"`
	}

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	return body.Code
}

func TestRerunRun(t *testing.T) {
	t.Parallel()

	t.Run("re-runs only the failed jobs of a failed run", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "77", Status: "completed", Conclusion: "failure"})

		rec := postAction(srv, run.ID, "rerun")

		require.Equal(t, http.StatusAccepted, rec.Code)
		require.Equal(t, "rerun", actor.do)
		require.True(t, actor.got.FailedOnly)
		require.Equal(t, "77", actor.got.ForgeRunID)
		require.Equal(t, "alrayyes/pipeline-analytics", actor.got.Identifier)
		require.Equal(t, "ghp_supersecrettoken1234", actor.got.Token)
	})

	t.Run("re-runs the whole run when it didn't fail", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "78", Status: "completed", Conclusion: "success"})

		require.Equal(t, http.StatusAccepted, postAction(srv, run.ID, "rerun").Code)
		require.False(t, actor.got.FailedOnly)
	})

	t.Run("a run that hasn't concluded is not actionable", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "79", Status: "in_progress"})

		rec := postAction(srv, run.ID, "rerun")

		require.Equal(t, http.StatusConflict, rec.Code)
		require.Equal(t, "not_actionable", errorCode(t, rec))
		require.Empty(t, actor.do, "the forge isn't asked")
	})

	t.Run("a token without Actions write is forbidden", func(t *testing.T) {
		t.Parallel()

		srv, repoID := actionTestServer(t, &stubRunActor{err: ingestion.ErrActionForbidden})
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "80", Status: "completed", Conclusion: "failure"})

		rec := postAction(srv, run.ID, "rerun")

		require.Equal(t, http.StatusForbidden, rec.Code)
		require.Equal(t, "forbidden", errorCode(t, rec))
		require.Contains(t, rec.Body.String(), "Actions", "the message names the missing permission")
		require.NotContains(t, rec.Body.String(), "ghp_supersecrettoken1234")
	})

	t.Run("a forge that didn't answer is a 502", func(t *testing.T) {
		t.Parallel()

		srv, repoID := actionTestServer(t, &stubRunActor{err: ingestion.ErrActionUnreachable})
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "81", Status: "completed", Conclusion: "failure"})

		rec := postAction(srv, run.ID, "rerun")

		require.Equal(t, http.StatusBadGateway, rec.Code)
		require.Equal(t, "unreachable", errorCode(t, rec))
	})

	t.Run("a forge with no actor is unsupported", func(t *testing.T) {
		t.Parallel()

		srv, repoID := actionTestServer(t, nil)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "82", Status: "completed", Conclusion: "failure"})

		rec := postAction(srv, run.ID, "rerun")

		require.Equal(t, http.StatusNotImplemented, rec.Code)
		require.Equal(t, "unsupported", errorCode(t, rec))
	})

	t.Run("an unexpected forge failure is a 500", func(t *testing.T) {
		t.Parallel()

		srv, repoID := actionTestServer(t, &stubRunActor{err: errStubReader})
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "83", Status: "completed", Conclusion: "failure"})

		require.Equal(t, http.StatusInternalServerError, postAction(srv, run.ID, "rerun").Code)
	})

	t.Run("an unknown run is a 404", func(t *testing.T) {
		t.Parallel()

		srv, _ := actionTestServer(t, &stubRunActor{})

		rec := postAction(srv, "missing", "rerun")

		require.Equal(t, http.StatusNotFound, rec.Code)
		require.Equal(t, "not_found", errorCode(t, rec))
	})

	t.Run("requires a session, and an API token isn't one", func(t *testing.T) {
		t.Parallel()

		srv, authStore, _ := actionTestServerWithAuthStore(t, &stubRunActor{})

		anon := httptest.NewRecorder()
		srv.ServeHTTP(anon, httptest.NewRequest(http.MethodPost, "/api/runs/x/rerun", nil))
		require.Equal(t, http.StatusUnauthorized, anon.Code)

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/runs/x/rerun", nil)
		req.Header.Set("Authorization", "Bearer "+raw)

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestCancelRun(t *testing.T) {
	t.Parallel()

	t.Run("cancels a run that's still running", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "90", Status: "in_progress"})

		require.Equal(t, http.StatusAccepted, postAction(srv, run.ID, "cancel").Code)
		require.Equal(t, "cancel", actor.do)
	})

	t.Run("a run that has concluded is not actionable", func(t *testing.T) {
		t.Parallel()

		actor := &stubRunActor{}
		srv, repoID := actionTestServer(t, actor)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "91", Status: "completed", Conclusion: "success"})

		rec := postAction(srv, run.ID, "cancel")

		require.Equal(t, http.StatusConflict, rec.Code)
		require.Empty(t, actor.do, "the forge isn't asked")
	})

	t.Run("a forge with no actor is unsupported", func(t *testing.T) {
		t.Parallel()

		srv, repoID := actionTestServer(t, nil)
		run := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "92", Status: "queued"})

		require.Equal(t, http.StatusNotImplemented, postAction(srv, run.ID, "cancel").Code)
	})

	t.Run("an API token isn't a session", func(t *testing.T) {
		t.Parallel()

		srv, authStore, _ := actionTestServerWithAuthStore(t, &stubRunActor{})

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/runs/x/cancel", nil)
		req.Header.Set("Authorization", "Bearer "+raw)

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

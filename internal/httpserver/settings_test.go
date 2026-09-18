package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/settings"
	settingssqlite "github.com/alrayyes/pipeline-analytics/internal/settings/sqlite"
	"github.com/stretchr/testify/require"
)

// newSettingsTestServer is a minimal httpserver.New() instance -- only the
// auth and settings deps this handler actually needs -- plus a ready
// session cookie.
func newSettingsTestServer(t *testing.T) testServer {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	authStore := authsqlite.NewStore(conn)
	ctx := context.Background()
	user, err := authStore.CreateUser(ctx, []byte("test-handle"), "admin")
	require.NoError(t, err)
	sessionID, err := authStore.CreateSession(ctx, user.ID)
	require.NoError(t, err)

	handler := httpserver.New(httpserver.Deps{
		Settings:  settings.NewService(settingssqlite.NewStore(conn)),
		AuthStore: authStore,
		Version:   "test-version",
		Assets:    testAssets,
	})

	cookie := &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	return testServer{Handler: handler, sessionCookie: cookie}
}

func TestSettingsGet(t *testing.T) {
	t.Parallel()

	t.Run("returns documented defaults for a fresh account", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/settings", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var got map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, settings.DefaultTheme, got["theme"])
		require.Equal(t, settings.DefaultForgeFilter, got["forgeFilter"])
		require.Equal(t, settings.DefaultPipelinesHealthFilter, got["pipelinesHealthFilter"])
		require.Equal(t, settings.DefaultPipelinesRepoSelector, got["pipelinesRepoSelector"])
		require.Equal(t, settings.DefaultPipelinesSortOrder, got["pipelinesSortOrder"])
	})

	t.Run("rejects an unauthenticated request", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestSettingsPatch(t *testing.T) {
	t.Parallel()

	t.Run("updates a setting, visible on the next GET", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		patchReq := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(`{"theme":"dark"}`))))
		patchRec := httptest.NewRecorder()
		srv.ServeHTTP(patchRec, patchReq)

		require.Equal(t, http.StatusOK, patchRec.Code)

		var patched map[string]string
		require.NoError(t, json.Unmarshal(patchRec.Body.Bytes(), &patched))
		require.Equal(t, "dark", patched["theme"])

		getReq := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/settings", nil))
		getRec := httptest.NewRecorder()
		srv.ServeHTTP(getRec, getReq)

		var got map[string]string
		require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &got))
		require.Equal(t, "dark", got["theme"])
	})

	t.Run("null reverts a setting to its default", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		first := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(`{"theme":"dark"}`))))
		srv.ServeHTTP(httptest.NewRecorder(), first)

		second := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(`{"theme":null}`))))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, second)

		require.Equal(t, http.StatusOK, rec.Code)

		var got map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, settings.DefaultTheme, got["theme"])
	})

	t.Run("resets the three Pipelines filters in one request", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		setup := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(
			`{"pipelinesHealthFilter":"healthy","pipelinesRepoSelector":"some-repo","pipelinesSortOrder":"lastRun"}`,
		))))
		srv.ServeHTTP(httptest.NewRecorder(), setup)

		reset := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(
			`{"pipelinesHealthFilter":null,"pipelinesRepoSelector":null,"pipelinesSortOrder":null}`,
		))))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, reset)

		require.Equal(t, http.StatusOK, rec.Code)

		var got map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, settings.DefaultPipelinesHealthFilter, got["pipelinesHealthFilter"])
		require.Equal(t, settings.DefaultPipelinesRepoSelector, got["pipelinesRepoSelector"])
		require.Equal(t, settings.DefaultPipelinesSortOrder, got["pipelinesSortOrder"])
	})

	t.Run("rejects an invalid enum value with 400, leaving the stored value unchanged", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		first := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(`{"theme":"dark"}`))))
		srv.ServeHTTP(httptest.NewRecorder(), first)

		bad := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(`{"theme":"not-a-theme"}`))))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, bad)

		require.Equal(t, http.StatusBadRequest, rec.Code)

		getReq := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/settings", nil))
		getRec := httptest.NewRecorder()
		srv.ServeHTTP(getRec, getReq)

		var got map[string]string
		require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &got))
		require.Equal(t, "dark", got["theme"])
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		req := srv.authenticated(httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(`not json`))))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects an unauthenticated request", func(t *testing.T) {
		t.Parallel()

		srv := newSettingsTestServer(t)

		req := httptest.NewRequest(http.MethodPatch, "/api/settings", bytes.NewReader([]byte(`{"theme":"dark"}`)))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

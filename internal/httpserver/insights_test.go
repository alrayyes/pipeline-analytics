package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	metricssqlite "github.com/alrayyes/pipeline-analytics/internal/metrics/sqlite"
	"github.com/stretchr/testify/require"
)

// newTestServerWithRateLimits is newTestServer, but with a GitHub
// fakeForgeClient that also reports the given rate-limit snapshots (keyed
// by token), wired as httpserver.Deps.GitHubRateLimits the same way
// cmd/pipeline-analytics/main.go wires the real client.
func newTestServerWithRateLimits(t *testing.T, rateLimits map[string]ingestion.RateLimitSnapshot) testServer {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	ingestionStore := ingestionsqlite.NewStore(conn, make([]byte, 32))
	githubClient := &fakeForgeClient{rateLimits: rateLimits}
	registrar := ingestion.NewRegistrar(ingestionStore, map[ingestion.Forge]ingestion.ForgeClient{
		ingestion.ForgeGitHub:  githubClient,
		ingestion.ForgeForgejo: &fakeForgeClient{},
	}, "https://example.com")

	authStore := authsqlite.NewStore(conn)
	ctx := context.Background()
	user, err := authStore.CreateUser(ctx, []byte("test-handle"), "admin")
	require.NoError(t, err)
	sessionID, err := authStore.CreateSession(ctx, user.ID)
	require.NoError(t, err)

	handler := httpserver.New(httpserver.Deps{
		Registrar:        registrar,
		IngestionStore:   ingestionStore,
		RunStore:         ingestionStore,
		Metrics:          metrics.NewService(metricssqlite.NewStore(conn)),
		AuthStore:        authStore,
		Version:          "test-version",
		Assets:           testAssets,
		GitHubRateLimits: githubClient,
	})

	cookie := &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	return testServer{Handler: handler, sessionCookie: cookie, runStore: ingestionStore}
}

// seedRepoWithToken registers a repo through the real API with an explicit
// identifier and token, returning its id.
func seedRepoWithToken(t *testing.T, srv testServer, identifier, token string) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"forge":      "github",
		"identifier": identifier,
		"token":      token,
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

func TestGitHubRateLimitInsights(t *testing.T) {
	t.Parallel()

	t.Run("groups repos sharing one token, with its rate-limit status", func(t *testing.T) {
		t.Parallel()

		resetAt := time.Unix(1700000000, 0).UTC()
		srv := newTestServerWithRateLimits(t, map[string]ingestion.RateLimitSnapshot{
			"ghp_shared": {Limit: 5000, Remaining: 4900, Used: 100, Resource: "core", ResetAt: resetAt},
		})
		seedRepoWithToken(t, srv, "alrayyes/repo-one", "ghp_shared")
		seedRepoWithToken(t, srv, "alrayyes/repo-two", "ghp_shared")

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/insights/github-rate-limit", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var groups []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &groups))
		require.Len(t, groups, 1) // one token, not one row per repo

		repos, _ := groups[0]["repos"].([]any)
		require.ElementsMatch(t, []any{"alrayyes/repo-one", "alrayyes/repo-two"}, repos)

		status, _ := groups[0]["status"].(map[string]any)
		require.InDelta(t, 5000, status["limit"], 0)
		require.InDelta(t, 4900, status["remaining"], 0)
		require.InDelta(t, 100, status["used"], 0)
		require.Equal(t, resetAt.Format(time.RFC3339), status["resetAt"])
	})

	t.Run("a token never observed on a real request has no status yet", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithRateLimits(t, nil)
		seedRepoWithToken(t, srv, "alrayyes/repo-one", "ghp_unseen")

		req := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/insights/github-rate-limit", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var groups []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &groups))
		require.Len(t, groups, 1)
		require.Nil(t, groups[0]["status"])
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithRateLimits(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/insights/github-rate-limit", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

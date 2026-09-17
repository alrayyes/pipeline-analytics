package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	metricssqlite "github.com/alrayyes/pipeline-analytics/internal/metrics/sqlite"
	"github.com/stretchr/testify/require"
)

type fakeForgeClient struct {
	err             error
	discoveredRepos []string
	discoverErr     error
	rateLimits      map[string]ingestion.RateLimitSnapshot
}

func (f *fakeForgeClient) RateLimitFor(token string) (ingestion.RateLimitSnapshot, bool) {
	snapshot, ok := f.rateLimits[token]

	return snapshot, ok
}

func (f *fakeForgeClient) CreateWebhook(context.Context, ingestion.CreateWebhookRequest) error {
	return f.err
}

func (f *fakeForgeClient) ListRecentRuns(context.Context, ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
	return ingestion.ListRunsResult{}, nil
}

func (f *fakeForgeClient) ListAccessibleRepos(context.Context, ingestion.ListAccessibleReposRequest) ([]string, error) {
	return f.discoveredRepos, f.discoverErr
}

var testAssets fs.FS = fstest.MapFS{"index.html": {Data: []byte("<html></html>")}}

// repoListResponse mirrors the handler's wrapped GET /api/repos response
// shape, kept loose (map per repo) since these tests only assert on a
// handful of fields.
type repoListResponse struct {
	Repos   []map[string]any `json:"repos"`
	HasMore bool             `json:"hasMore"`
}

// testServer is an httpserver.New() instance plus a ready-made session
// cookie for tests that need to call a gated endpoint.
type testServer struct {
	http.Handler
	sessionCookie *http.Cookie
	runStore      ingestion.RunStore
}

// authenticated clones req with the test server's session cookie attached.
func (s testServer) authenticated(req *http.Request) *http.Request {
	req.AddCookie(s.sessionCookie)

	return req
}

func newTestServer(t *testing.T, forgeErr error) testServer {
	t.Helper()

	return newTestServerWithAssets(t, forgeErr, testAssets)
}

func newTestServerWithAssets(t *testing.T, forgeErr error, assets fs.FS) testServer {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	ingestionStore := ingestionsqlite.NewStore(conn, make([]byte, 32))
	registrar := ingestion.NewRegistrar(ingestionStore, map[ingestion.Forge]ingestion.ForgeClient{
		ingestion.ForgeGitHub:  &fakeForgeClient{err: forgeErr},
		ingestion.ForgeForgejo: &fakeForgeClient{err: forgeErr},
	}, "https://example.com")

	authStore := authsqlite.NewStore(conn)
	ctx := context.Background()
	user, err := authStore.CreateUser(ctx, []byte("test-handle"), "admin")
	require.NoError(t, err)
	sessionID, err := authStore.CreateSession(ctx, user.ID)
	require.NoError(t, err)

	handler := httpserver.New(httpserver.Deps{
		Registrar:      registrar,
		IngestionStore: ingestionStore,
		RunStore:       ingestionStore,
		Metrics:        metrics.NewService(metricssqlite.NewStore(conn)),
		AuthStore:      authStore,
		Version:        "test-version",
		Assets:         assets,
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

func newTestServerWithDiscovery(t *testing.T, discoveredRepos []string, discoverErr error) testServer {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	ingestionStore := ingestionsqlite.NewStore(conn, make([]byte, 32))
	client := &fakeForgeClient{discoveredRepos: discoveredRepos, discoverErr: discoverErr}
	registrar := ingestion.NewRegistrar(ingestionStore, map[ingestion.Forge]ingestion.ForgeClient{
		ingestion.ForgeGitHub:  client,
		ingestion.ForgeForgejo: client,
	}, "https://example.com")

	authStore := authsqlite.NewStore(conn)
	ctx := context.Background()
	user, err := authStore.CreateUser(ctx, []byte("test-handle"), "admin")
	require.NoError(t, err)
	sessionID, err := authStore.CreateSession(ctx, user.ID)
	require.NoError(t, err)

	handler := httpserver.New(httpserver.Deps{
		Registrar:      registrar,
		IngestionStore: ingestionStore,
		RunStore:       ingestionStore,
		Metrics:        metrics.NewService(metricssqlite.NewStore(conn)),
		AuthStore:      authStore,
		Version:        "test-version",
		Assets:         testAssets,
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

func TestReposDiscover(t *testing.T) {
	t.Parallel()

	t.Run("returns the repos the token can access", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithDiscovery(t, []string{"alrayyes/pipeline-analytics", "alrayyes/dotfiles"}, nil)

		body, err := json.Marshal(map[string]string{
			"forge": "github",
			"token": "ghp_supersecrettoken1234",
		})
		require.NoError(t, err)

		req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos/discover", bytes.NewReader(body)))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var got []string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, []string{"alrayyes/pipeline-analytics", "alrayyes/dotfiles"}, got)
	})

	t.Run("surfaces a forge error as 502", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithDiscovery(t, nil, errPermissionDenied)

		body, err := json.Marshal(map[string]string{ //nolint:gosec // test fixture value, not a real credential
			"forge": "github",
			"token": "ghp_badtoken",
		})
		require.NoError(t, err)

		req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos/discover", bytes.NewReader(body)))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadGateway, rec.Code)
	})

	t.Run("rejects a request missing the token", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithDiscovery(t, nil, nil)

		req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos/discover", bytes.NewReader([]byte(`{"forge":"github"}`))))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects an unauthenticated request", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithDiscovery(t, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/repos/discover", bytes.NewReader([]byte(`{}`)))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestReposRegisterAndList(t *testing.T) {
	t.Parallel()

	t.Run("registers a repo and returns it masked, active on webhook success", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		body, err := json.Marshal(map[string]string{
			"forge":      "github",
			"identifier": "alrayyes/pipeline-analytics",
			"token":      "ghp_supersecrettoken1234",
		})
		require.NoError(t, err)

		req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body)))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)

		var got map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, "active", got["ingestionStatus"])
		require.Equal(t, "****1234", got["tokenMasked"])
		require.NotContains(t, rec.Body.String(), "supersecrettoken")
	})

	t.Run("webhook failure registers the repo as degraded, not an error response", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, errPermissionDenied)

		body, err := json.Marshal(map[string]string{
			"forge":      "github",
			"identifier": "alrayyes/pipeline-analytics",
			"token":      "ghp_supersecrettoken1234",
		})
		require.NoError(t, err)

		req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body)))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)

		var got map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, "degraded", got["ingestionStatus"])
		require.Contains(t, got["ingestionStatusReason"], "permission denied")
	})

	t.Run("rejects a request missing required fields", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader([]byte(`{"forge":"github"}`))))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects an unauthenticated request", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("list returns every registered repo", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		body, err := json.Marshal(map[string]string{
			"forge":      "github",
			"identifier": "alrayyes/pipeline-analytics",
			"token":      "ghp_supersecrettoken1234",
		})
		require.NoError(t, err)

		postReq := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body)))
		srv.ServeHTTP(httptest.NewRecorder(), postReq)

		getReq := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/repos", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, getReq)

		require.Equal(t, http.StatusOK, rec.Code)

		var got repoListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.False(t, got.HasMore)
		require.Len(t, got.Repos, 1)
	})

	t.Run("list filters by the forge query param", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		register := func(forge, identifier string) {
			t.Helper()

			body, err := json.Marshal(map[string]string{
				"forge":      forge,
				"identifier": identifier,
				"token":      "ghp_supersecrettoken1234",
			})
			require.NoError(t, err)

			req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body)))
			srv.ServeHTTP(httptest.NewRecorder(), req)
		}
		register("github", "alrayyes/pipeline-analytics")
		register("forgejo", "alrayyes/dotfiles")

		getReq := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/repos?forge=github", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, getReq)

		require.Equal(t, http.StatusOK, rec.Code)

		var got repoListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got.Repos, 1)
		require.Equal(t, "github", got.Repos[0]["forge"])
	})

	t.Run("list paginates with limit and offset", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		register := func(identifier string) {
			t.Helper()

			body, err := json.Marshal(map[string]string{
				"forge":      "github",
				"identifier": identifier,
				"token":      "ghp_supersecrettoken1234",
			})
			require.NoError(t, err)

			req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body)))
			srv.ServeHTTP(httptest.NewRecorder(), req)
		}
		register("alrayyes/repo-0")
		register("alrayyes/repo-1")
		register("alrayyes/repo-2")

		firstPage := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/repos?limit=2&offset=0", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, firstPage)

		require.Equal(t, http.StatusOK, rec.Code)

		var got repoListResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.True(t, got.HasMore)
		require.Len(t, got.Repos, 2)

		secondPage := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/repos?limit=2&offset=2", nil))
		rec = httptest.NewRecorder()
		srv.ServeHTTP(rec, secondPage)

		require.Equal(t, http.StatusOK, rec.Code)

		got = repoListResponse{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.False(t, got.HasMore)
		require.Len(t, got.Repos, 1)
	})
}

func TestReposUntrack(t *testing.T) {
	t.Parallel()

	t.Run("untracking a repo removes it from the list", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		body, err := json.Marshal(map[string]string{
			"forge":      "github",
			"identifier": "alrayyes/pipeline-analytics",
			"token":      "ghp_supersecrettoken1234",
		})
		require.NoError(t, err)

		postReq := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body)))
		postRec := httptest.NewRecorder()
		srv.ServeHTTP(postRec, postReq)

		var created map[string]any
		require.NoError(t, json.Unmarshal(postRec.Body.Bytes(), &created))
		id, ok := created["id"].(string)
		require.True(t, ok)

		delReq := srv.authenticated(httptest.NewRequest(http.MethodDelete, "/api/repos/"+id, nil))
		delRec := httptest.NewRecorder()
		srv.ServeHTTP(delRec, delReq)

		require.Equal(t, http.StatusNoContent, delRec.Code)

		getReq := srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/repos", nil))
		getRec := httptest.NewRecorder()
		srv.ServeHTTP(getRec, getReq)

		var got repoListResponse
		require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &got))
		require.Empty(t, got.Repos)
	})

	t.Run("untracking an unknown repo returns 404", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := srv.authenticated(httptest.NewRequest(http.MethodDelete, "/api/repos/does-not-exist", nil))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("rejects an unauthenticated request", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/repos/some-id", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

var errPermissionDenied = permissionDeniedError{}

type permissionDeniedError struct{}

func (permissionDeniedError) Error() string { return "permission denied" }

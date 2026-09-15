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

	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/stretchr/testify/require"
)

type fakeForgeClient struct {
	err error
}

func (f *fakeForgeClient) CreateWebhook(context.Context, ingestion.CreateWebhookRequest) error {
	return f.err
}

var testAssets fs.FS = fstest.MapFS{"index.html": {Data: []byte("<html></html>")}}

func newTestServer(t *testing.T, forgeErr error) http.Handler {
	t.Helper()

	return newTestServerWithAssets(t, forgeErr, testAssets)
}

func newTestServerWithAssets(t *testing.T, forgeErr error, assets fs.FS) http.Handler {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	store := sqlite.NewStore(conn, make([]byte, 32))
	registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
		ingestion.ForgeGitHub: &fakeForgeClient{err: forgeErr},
	}, "https://example.com")

	return httpserver.New(registrar, store, "test-version", assets)
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

		req := httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body))
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

		req := httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body))
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

		req := httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader([]byte(`{"forge":"github"}`)))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
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

		postReq := httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body))
		srv.ServeHTTP(httptest.NewRecorder(), postReq)

		getReq := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, getReq)

		require.Equal(t, http.StatusOK, rec.Code)

		var got []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got, 1)
	})
}

var errPermissionDenied = permissionDeniedError{}

type permissionDeniedError struct{}

func (permissionDeniedError) Error() string { return "permission denied" }

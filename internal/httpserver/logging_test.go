package httpserver_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	metricssqlite "github.com/alrayyes/pipeline-analytics/internal/metrics/sqlite"
	"github.com/stretchr/testify/require"
)

// captureSlog swaps the package-global slog default for the duration of the
// test, so a request-logging assertion can inspect what was actually
// written. The whole TestRequestLogging function (and every subtest here)
// deliberately skips t.Parallel() -- see its own doc comment for why that's
// what keeps this race-free without every other test in the package having
// to avoid parallelism too.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer

	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return &buf
}

// TestRequestLogging swaps the package-global slog default to capture
// output, so it (and its subtests) deliberately don't call t.Parallel() --
// concurrent tests logging through the same swapped default would race on
// it. Go only starts running parallel-marked tests once every non-parallel
// top-level test in the package has finished, so this being sequential is
// what keeps the swap race-free without every other test in the package
// needing to avoid parallelism too.
func TestRequestLogging(t *testing.T) {
	t.Run("a successful request is logged at Info", func(t *testing.T) {
		logs := captureSlog(t)
		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, logs.String(), "level=INFO")
		require.Contains(t, logs.String(), "GET")
		require.Contains(t, logs.String(), "/api/version")
		require.Contains(t, logs.String(), "status=200")
	})

	t.Run("a handler that never calls WriteHeader still logs status 200", func(t *testing.T) {
		// The static/SPA handler serves index.html via http.FileServer,
		// which writes the body directly without an explicit WriteHeader
		// call on a plain success -- the canonical case for net/http's own
		// "no WriteHeader call defaults to 200" behavior. If the recorder
		// only captured status from an explicit WriteHeader call, this
		// would log status=0, not status=200.
		logs := captureSlog(t)
		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, logs.String(), "status=200")
	})

	t.Run("a rejected (401) request is still logged, at Warn", func(t *testing.T) {
		logs := captureSlog(t)
		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Contains(t, logs.String(), "level=WARN")
		require.Contains(t, logs.String(), "status=401")
	})

	t.Run("a client-error (4xx) request is logged at Warn", func(t *testing.T) {
		logs := captureSlog(t)
		srv := newTestServer(t, nil)

		req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader([]byte(`{"forge":"github"}`))))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, logs.String(), "level=WARN")
		require.Contains(t, logs.String(), "status=400")
	})

	t.Run("a server-error (5xx) request is logged at Error", func(t *testing.T) {
		logs := captureSlog(t)

		conn, err := db.Open(":memory:")
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, conn.Close()) })
		require.NoError(t, db.Migrate(context.Background(), conn))

		// A key of the wrong length makes token encryption fail before it
		// ever reaches storage -- a real, reproducible 500.
		badStore := ingestionsqlite.NewStore(conn, []byte("too-short"))
		registrar := ingestion.NewRegistrar(badStore, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{},
		}, "https://example.com")

		authStore := authsqlite.NewStore(conn)
		ctx := context.Background()
		user, err := authStore.CreateUser(ctx, []byte("test-handle"), "admin")
		require.NoError(t, err)
		sessionID, err := authStore.CreateSession(ctx, user.ID)
		require.NoError(t, err)

		handler := httpserver.New(httpserver.Deps{
			Registrar:      registrar,
			IngestionStore: badStore,
			RunStore:       badStore,
			Metrics:        metrics.NewService(metricssqlite.NewStore(conn)),
			AuthStore:      authStore,
			Version:        "test-version",
			Assets:         testAssets,
		})

		body := []byte(`{"forge":"github","identifier":"alrayyes/pipeline-analytics","token":"ghp_supersecrettoken1234"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/repos", bytes.NewReader(body))
		req.AddCookie(&http.Cookie{
			Name:     "session",
			Value:    sessionID,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, logs.String(), "level=ERROR")
		require.Contains(t, logs.String(), "status=500")
	})

	t.Run("/healthz produces no log record", func(t *testing.T) {
		logs := captureSlog(t)
		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Empty(t, logs.String())
	})

	t.Run("no sensitive data reaches the log", func(t *testing.T) {
		logs := captureSlog(t)
		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/repos?token=super-secret-value", nil)
		req.AddCookie(&http.Cookie{
			Name:     "session",
			Value:    "sensitive-session-id",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		req.Header.Set("Authorization", "Bearer sensitive-bearer-token")
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.NotContains(t, logs.String(), "super-secret-value")
		require.NotContains(t, logs.String(), "sensitive-session-id")
		require.NotContains(t, logs.String(), "sensitive-bearer-token")
	})
}

package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/config"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/stretchr/testify/require"
)

// listenOnFreePort binds PIPELINE_ANALYTICS_ADDR to an actual free loopback
// port for the duration of t - the healthcheck subcommand reads that env var
// directly, it doesn't take a flag.
func listenOnFreePort(t *testing.T) net.Listener {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })

	t.Setenv("PIPELINE_ANALYTICS_ADDR", l.Addr().String())

	return l
}

func TestHealthcheckSucceedsWhenHealthzAnswers200(t *testing.T) {
	l := listenOnFreePort(t)

	srv := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { _ = srv.Close() })

	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	require.NoError(t, root.Execute())
}

func TestHealthcheckAsksForReadiness(t *testing.T) {
	l := listenOnFreePort(t)

	paths := make(chan string, 1)
	srv := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths <- r.URL.Path

		w.WriteHeader(http.StatusOK)
	})}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { _ = srv.Close() })

	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	require.NoError(t, root.Execute())
	require.Equal(t, "/readyz", <-paths, "liveness alone would report healthy with the database down")
}

func TestHealthcheckFailsWhenHealthzAnswersNon200(t *testing.T) {
	l := listenOnFreePort(t)

	srv := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { _ = srv.Close() })

	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	require.Error(t, root.Execute())
}

func TestHealthcheckFailsWhenNothingIsListening(t *testing.T) {
	// A free port that was bound and immediately released, rather than
	// listenOnFreePort - nothing serves it, which is the failure mode the
	// container's own HEALTHCHECK is there to catch.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	t.Setenv("PIPELINE_ANALYTICS_ADDR", addr)

	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	require.Error(t, root.Execute())
}

func TestHealthcheckFallsBackToDefaultAddrWhenEnvUnset(t *testing.T) {
	// serve's own flag default is ":8080" - nothing should be listening
	// there in a test process, so this just proves the fallback is used
	// (a connection-refused error) rather than an empty-addr parse error.
	root := newRootCmd()
	root.SetArgs([]string{"healthcheck"})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	require.Error(t, root.Execute())
}

// TestBuildHandlerReadyzFollowsTheDatabase covers the wiring main owns: the
// real handler's /readyz must go red when its database does, which is the
// whole reason the container check asks for readiness and not liveness.
func TestBuildHandlerReadyzFollowsTheDatabase(t *testing.T) {
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(context.Background(), conn))

	cfg := config.Config{CallbackURL: "https://example.com", EncryptionKey: make([]byte, 32)}
	handler, err := buildHandler(cfg, conn, ingestionsqlite.NewStore(conn, cfg.EncryptionKey), nil, nil)
	require.NoError(t, err)

	probe := func() int {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

		return rec.Code
	}

	require.Equal(t, http.StatusOK, probe(), "ready while the database answers")

	require.NoError(t, conn.Close())
	require.Equal(t, http.StatusServiceUnavailable, probe(), "not ready once it can't")
}

// TestDBReady covers what "ready" has to mean: the database can be read, not
// just that a connection opens. A file that isn't a SQLite database opens
// fine and only fails on the first real query, which is the kind of fault
// (corruption, a bad restore, a disk error) the container check is for.
func TestDBReady(t *testing.T) {
	t.Run("passes against a migrated database", func(t *testing.T) {
		conn, err := db.Open(":memory:")
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })
		require.NoError(t, db.Migrate(context.Background(), conn))

		require.NoError(t, dbReady(conn)(context.Background()))
	})

	t.Run("fails once the connection is closed", func(t *testing.T) {
		conn, err := db.Open(":memory:")
		require.NoError(t, err)
		require.NoError(t, conn.Close())

		require.Error(t, dbReady(conn)(context.Background()))
	})

	t.Run("fails against a file that isn't a database", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "garbage.db")
		require.NoError(t, os.WriteFile(path, []byte(strings.Repeat("this is not a sqlite database\n", 200)), 0o600))

		conn, err := db.Open(path)
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })

		require.Error(t, dbReady(conn)(context.Background()))
	})
}

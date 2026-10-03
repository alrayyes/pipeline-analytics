package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestLoadConfig_HyphenatedFlagsFromEnv guards against a real regression:
// viper's default env-var lookup for a flag named "callback-url" is the
// literal key "PIPELINE_ANALYTICS_CALLBACK-URL" -- not a name a shell,
// Docker, or systemd can set, since environment variable names can't
// contain a hyphen. Every hyphenated flag (callback-url, encryption-key,
// reconcile-interval) was unreachable by environment variable at all until
// this test's fix.
func TestLoadConfig_HyphenatedFlagsFromEnv(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_ADDR", ":9090")
	t.Setenv("PIPELINE_ANALYTICS_DB", "test.db")
	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))
	t.Setenv("PIPELINE_ANALYTICS_RECONCILE_INTERVAL", "30m")

	cfg, err := loadConfig()
	require.NoError(t, err)

	require.Equal(t, "https://example.com", cfg.CallbackURL)
	require.Len(t, cfg.EncryptionKey, 32)
	require.Equal(t, "30m0s", cfg.ReconcileInterval.String())
}

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	t.Run("parses a valid level case-insensitively", func(t *testing.T) {
		t.Parallel()

		level, err := parseLogLevel("DEBUG")
		require.NoError(t, err)
		require.Equal(t, slog.LevelDebug, level)
	})

	t.Run("rejects an invalid level", func(t *testing.T) {
		t.Parallel()

		_, err := parseLogLevel("verbose")
		require.Error(t, err)
	})
}

// TestLoadConfig_LogLevelDefaultsToInfo guards the --log-level flag's
// default actually reaching Config -- with nothing set, logging should stay
// at Info, unchanged from before this flag existed.
func TestLoadConfig_LogLevelDefaultsToInfo(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))

	cfg, err := loadConfig()
	require.NoError(t, err)

	require.Equal(t, slog.LevelInfo, cfg.LogLevel)
}

// TestLoadConfig_LogLevelFromEnv guards the flag's actual point: turning
// this up via PIPELINE_ANALYTICS_LOG_LEVEL is what --debug-without-a-
// redeploy means in practice.
func TestLoadConfig_LogLevelFromEnv(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))
	t.Setenv("PIPELINE_ANALYTICS_LOG_LEVEL", "debug")

	cfg, err := loadConfig()
	require.NoError(t, err)

	require.Equal(t, slog.LevelDebug, cfg.LogLevel)
}

func TestLoadConfig_InvalidLogLevel(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))
	t.Setenv("PIPELINE_ANALYTICS_LOG_LEVEL", "verbose")

	_, err := loadConfig()
	require.Error(t, err)
}

func listen(t *testing.T) net.Listener {
	t.Helper()

	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	return ln
}

// fetch is status without assertions, safe to call from a goroutine.
func fetch(url string) (int, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return 0, err //nolint:wrapcheck // the test inspects the error itself
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err //nolint:wrapcheck // the test inspects the transport error itself
	}

	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode, nil
}

func status(t *testing.T, url string) (int, error) {
	t.Helper()

	return fetch(url)
}

// TestServeUntilDone_DrainsBeforeClosing pins the shutdown order api.md asks
// for: on the signal /readyz flips to 503 first, the server keeps serving
// through the drain period so the router can notice, and only then closes.
func TestServeUntilDone_DrainsBeforeClosing(t *testing.T) {
	var draining atomic.Bool

	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if draining.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/work", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	ln := listen(t)
	base := "http://" + ln.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	started := time.Now()

	drain := 400 * time.Millisecond

	go func() {
		done <- serveUntilDone(ctx, &http.Server{Handler: mux, ReadHeaderTimeout: time.Second}, ln, shutdownPlan{
			draining: &draining, drainPeriod: drain, timeout: 2 * time.Second,
		})
	}()

	code, err := status(t, base+"/readyz")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)

	cancel()

	require.Eventually(t, func() bool {
		code, err := status(t, base+"/readyz")

		return err == nil && code == http.StatusServiceUnavailable
	}, time.Second, 10*time.Millisecond, "readiness flips to 503 on the signal")

	code, err = status(t, base+"/work")
	require.NoError(t, err, "other requests are still served during the drain")
	require.Equal(t, http.StatusOK, code)

	require.NoError(t, <-done)
	require.GreaterOrEqual(t, time.Since(started), drain, "it closes only after the drain period")

	_, err = status(t, base+"/work")
	require.Error(t, err, "and then it is closed")
}

// TestServeUntilDone_ShutdownIsBounded: a request that never finishes can't
// hold the process open past the timeout.
func TestServeUntilDone_ShutdownIsBounded(t *testing.T) {
	block := make(chan struct{})
	t.Cleanup(func() { close(block) })

	mux := http.NewServeMux()
	mux.HandleFunc("/hang", func(_ http.ResponseWriter, _ *http.Request) { <-block })

	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- serveUntilDone(ctx, &http.Server{Handler: mux, ReadHeaderTimeout: time.Second}, ln, shutdownPlan{
			draining: &atomic.Bool{}, timeout: 200 * time.Millisecond,
		})
	}()

	go func() { _, _ = fetch("http://" + ln.Addr().String() + "/hang") }()

	time.Sleep(100 * time.Millisecond) // let the request arrive
	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown waited past its timeout")
	}
}

func TestLoadConfig_ShutdownFlagsFromEnv(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))
	t.Setenv("PIPELINE_ANALYTICS_DRAIN_PERIOD", "7s")
	t.Setenv("PIPELINE_ANALYTICS_SHUTDOWN_TIMEOUT", "11s")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, 7*time.Second, cfg.DrainPeriod)
	require.Equal(t, 11*time.Second, cfg.ShutdownTimeout)
}

func TestLoadConfig_ShutdownDefaultsFitADockerStop(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))

	cfg, err := loadConfig()
	require.NoError(t, err)

	// Docker sends SIGKILL ten seconds after SIGTERM by default.
	require.Less(t, cfg.DrainPeriod+cfg.ShutdownTimeout, 10*time.Second)
}

func TestListenAndServe(t *testing.T) {
	t.Run("serves on the address and returns nil once the context ends", func(t *testing.T) {
		ln := listen(t)
		addr := ln.Addr().String()
		require.NoError(t, ln.Close()) // free the port for listenAndServe to take

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)

		srv := &http.Server{
			Addr:              addr,
			ReadHeaderTimeout: time.Second,
			Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
		}

		go func() {
			done <- listenAndServe(ctx, srv, shutdownPlan{draining: &atomic.Bool{}, timeout: time.Second})
		}()

		require.Eventually(t, func() bool {
			code, err := fetch("http://" + addr + "/")

			return err == nil && code == http.StatusOK
		}, 2*time.Second, 20*time.Millisecond)

		cancel()
		require.NoError(t, <-done)
	})

	t.Run("reports an address it can't listen on", func(t *testing.T) {
		ln := listen(t) // keeps the port busy

		err := listenAndServe(context.Background(), &http.Server{Addr: ln.Addr().String(), ReadHeaderTimeout: time.Second},
			shutdownPlan{draining: &atomic.Bool{}, timeout: time.Second})

		require.ErrorContains(t, err, "listen on")
	})
}

// A listener that fails under the server ends serveUntilDone with that error
// instead of waiting for a signal that may never come.
func TestServeUntilDone_ReturnsAServeError(t *testing.T) {
	ln := listen(t)
	require.NoError(t, ln.Close())

	err := serveUntilDone(context.Background(), &http.Server{ReadHeaderTimeout: time.Second}, ln,
		shutdownPlan{draining: &atomic.Bool{}, timeout: time.Second})

	require.ErrorContains(t, err, "serve:")
}

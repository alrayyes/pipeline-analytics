package httpserver_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/stretchr/testify/require"
)

var (
	errDiskIO    = errors.New("unable to open /var/lib/pipeline-analytics/data.db: disk I/O error")
	errProbeDown = errors.New("down")
)

func readyzRequest(t *testing.T, deps httpserver.Deps) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	httpserver.New(deps).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	return rec
}

// TestReadyz swaps the package-global slog default (captureSlog), so like
// TestRequestLogging it and its subtests deliberately don't call t.Parallel().
func TestReadyz(t *testing.T) {
	t.Run("answers 200 when the readiness probe passes, with no session", func(t *testing.T) {
		rec := readyzRequest(t, httpserver.Deps{Ready: func(context.Context) error { return nil }})

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "ok", rec.Body.String())
	})

	t.Run("answers 503 when the probe fails, without leaking the cause", func(t *testing.T) {
		logs := captureSlog(t)
		rec := readyzRequest(t, httpserver.Deps{Ready: func(context.Context) error {
			return errDiskIO
		}})

		require.Equal(t, http.StatusServiceUnavailable, rec.Code)
		require.JSONEq(t, `{"code":"not_ready","message":"not ready"}`, rec.Body.String())
		require.NotContains(t, rec.Body.String(), "data.db")
		require.Contains(t, logs.String(), "data.db", "the cause belongs in the log")
		require.Contains(t, logs.String(), "level=ERROR")
	})

	t.Run("is ready when no probe is configured", func(t *testing.T) {
		require.Equal(t, http.StatusOK, readyzRequest(t, httpserver.Deps{}).Code)
	})

	t.Run("gives the probe a deadline, so a hung database fails instead of hanging the check", func(t *testing.T) {
		var remaining time.Duration

		readyzRequest(t, httpserver.Deps{Ready: func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			require.True(t, ok, "the probe's context has no deadline")

			remaining = time.Until(deadline)

			return nil
		}})

		require.Positive(t, remaining)
		require.LessOrEqual(t, remaining, 5*time.Second)
	})

	t.Run("a probe that outlives its deadline gives 503", func(t *testing.T) {
		captureSlog(t)

		rec := readyzRequest(t, httpserver.Deps{Ready: func(ctx context.Context) error {
			<-ctx.Done()

			return ctx.Err()
		}})

		require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("liveness is unaffected by a failing probe", func(t *testing.T) {
		rec := httptest.NewRecorder()
		httpserver.New(httpserver.Deps{Ready: func(context.Context) error { return errProbeDown }}).
			ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("a passing check produces no request log record", func(t *testing.T) {
		logs := captureSlog(t)
		readyzRequest(t, httpserver.Deps{Ready: func(context.Context) error { return nil }})

		require.Empty(t, logs.String())
	})
}

func serveReadyz(h http.Handler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	return rec
}

func TestReadyz_Draining(t *testing.T) {
	var draining atomic.Bool

	var probes atomic.Int32

	h := httpserver.New(httpserver.Deps{
		Draining: &draining,
		Ready: func(context.Context) error {
			probes.Add(1)

			return nil
		},
	})

	require.Equal(t, http.StatusOK, serveReadyz(h).Code)

	draining.Store(true)

	rec := serveReadyz(h)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.JSONEq(t, `{"code":"not_ready","message":"not ready"}`, rec.Body.String())

	t.Run("a draining instance doesn't probe its dependency", func(t *testing.T) {
		before := probes.Load()
		serveReadyz(h)
		require.Equal(t, before, probes.Load())
	})

	t.Run("liveness stays up while draining", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		require.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestReadyz_CachesTheProbe(t *testing.T) {
	var probes atomic.Int32

	h := httpserver.New(httpserver.Deps{
		ReadyCacheTTL: time.Hour,
		Ready: func(context.Context) error {
			probes.Add(1)

			return nil
		},
	})

	for range 5 {
		require.Equal(t, http.StatusOK, serveReadyz(h).Code)
	}

	require.Equal(t, int32(1), probes.Load(), "polling inside the window reuses the result")
}

func TestReadyz_CacheExpiresSoARecoveryShows(t *testing.T) {
	var down atomic.Bool

	down.Store(true)

	h := httpserver.New(httpserver.Deps{
		ReadyCacheTTL: 20 * time.Millisecond,
		Ready: func(context.Context) error {
			if down.Load() {
				return errProbeDown
			}

			return nil
		},
	})

	captureSlog(t)
	require.Equal(t, http.StatusServiceUnavailable, serveReadyz(h).Code)

	down.Store(false)

	require.Eventually(t, func() bool { return serveReadyz(h).Code == http.StatusOK }, time.Second, 10*time.Millisecond)
}

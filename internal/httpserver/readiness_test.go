package httpserver_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

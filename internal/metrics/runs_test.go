package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestParseRunStatus(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"all", "failed", "running", "success"} {
		got, err := metrics.ParseRunStatus(raw)
		require.NoError(t, err)
		require.Equal(t, metrics.RunStatus(raw), got)
	}

	t.Run("empty means all, the contract's default", func(t *testing.T) {
		t.Parallel()

		got, err := metrics.ParseRunStatus("")
		require.NoError(t, err)
		require.Equal(t, metrics.RunStatusAll, got)
	})

	t.Run("anything else is rejected, not coerced", func(t *testing.T) {
		t.Parallel()

		for _, raw := range []string{"FAILED", "passed", "cancelled", "1"} {
			_, err := metrics.ParseRunStatus(raw)
			require.ErrorIs(t, err, metrics.ErrInvalidRunStatus, raw)
		}
	})
}

func TestRunEntry_DurationSeconds(t *testing.T) {
	t.Parallel()

	started := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	completed := started.Add(6*time.Minute + 42*time.Second)

	t.Run("is the span between start and completion", func(t *testing.T) {
		t.Parallel()

		secs, ok := metrics.RunEntry{StartedAt: &started, CompletedAt: &completed}.DurationSeconds()
		require.True(t, ok)
		require.InDelta(t, 402, secs, 1e-9)
	})

	t.Run("is absent while the run is still going", func(t *testing.T) {
		t.Parallel()

		_, ok := metrics.RunEntry{StartedAt: &started}.DurationSeconds()
		require.False(t, ok)
	})

	t.Run("is absent for a run that never started", func(t *testing.T) {
		t.Parallel()

		_, ok := metrics.RunEntry{CompletedAt: &completed}.DurationSeconds()
		require.False(t, ok)
	})
}

func TestService_ListRuns(t *testing.T) {
	t.Parallel()

	t.Run("passes the filter through and returns the page with its has-more flag", func(t *testing.T) {
		t.Parallel()

		store := &fakeStore{runList: []metrics.RunEntry{{ID: "r1"}, {ID: "r2"}}, runListHasMore: true}

		runs, hasMore, err := metrics.NewService(store).ListRuns(context.Background(), metrics.RunListFilter{Status: metrics.RunStatusFailed, Limit: 2})
		require.NoError(t, err)
		require.True(t, hasMore)
		require.Len(t, runs, 2)
		require.Equal(t, metrics.RunStatusFailed, store.gotRunListFilter.Status)
		require.Equal(t, 2, store.gotRunListFilter.Limit)
	})
}

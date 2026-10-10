package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestStore_Quarantine(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

	key := func(f testFixture) metrics.QuarantineKey {
		return metrics.QuarantineKey{Pipeline: metrics.PipelineRef{RepoID: f.repo.ID, Name: "CI"}, Step: "test"}
	}

	t.Run("a quarantined step is active for 30 days from the mark", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t)
		ctx := context.Background()

		got, err := f.metrics.QuarantineStep(ctx, key(f), "waits on the shared database", now)
		require.NoError(t, err)
		require.Equal(t, "waits on the shared database", got.Note)
		require.True(t, got.QuarantinedAt.Equal(now))
		require.True(t, got.ExpiresAt.Equal(now.Add(metrics.QuarantineDuration)))

		active, err := f.metrics.ActiveQuarantines(ctx, now.Add(29*24*time.Hour))
		require.NoError(t, err)
		require.Len(t, active, 1)
		require.Equal(t, "waits on the shared database", active[key(f)].Note)
	})

	t.Run("an expired mark behaves as if it was never set", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t)
		ctx := context.Background()

		_, err := f.metrics.QuarantineStep(ctx, key(f), "", now)
		require.NoError(t, err)

		active, err := f.metrics.ActiveQuarantines(ctx, now.Add(31*24*time.Hour))
		require.NoError(t, err)
		require.Empty(t, active)
	})

	t.Run("the instant of expiry is already inactive", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t)
		ctx := context.Background()

		_, err := f.metrics.QuarantineStep(ctx, key(f), "", now)
		require.NoError(t, err)

		active, err := f.metrics.ActiveQuarantines(ctx, now.Add(metrics.QuarantineDuration))
		require.NoError(t, err)
		require.Empty(t, active)
	})

	t.Run("marking again renews it and replaces the note", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t)
		ctx := context.Background()

		_, err := f.metrics.QuarantineStep(ctx, key(f), "first", now)
		require.NoError(t, err)

		later := now.Add(20 * 24 * time.Hour)
		_, err = f.metrics.QuarantineStep(ctx, key(f), "second", later)
		require.NoError(t, err)

		active, err := f.metrics.ActiveQuarantines(ctx, later.Add(25*24*time.Hour))
		require.NoError(t, err)
		require.Len(t, active, 1)
		require.Equal(t, "second", active[key(f)].Note)
		require.True(t, active[key(f)].ExpiresAt.Equal(later.Add(metrics.QuarantineDuration)))
	})

	t.Run("un-quarantining removes it, and is fine when there is none", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t)
		ctx := context.Background()

		require.NoError(t, f.metrics.UnquarantineStep(ctx, key(f)))

		_, err := f.metrics.QuarantineStep(ctx, key(f), "", now)
		require.NoError(t, err)
		require.NoError(t, f.metrics.UnquarantineStep(ctx, key(f)))

		active, err := f.metrics.ActiveQuarantines(ctx, now)
		require.NoError(t, err)
		require.Empty(t, active)
	})

	t.Run("a step is one name within one pipeline", func(t *testing.T) {
		t.Parallel()

		f := newFixture(t)
		ctx := context.Background()
		other := metrics.QuarantineKey{Pipeline: metrics.PipelineRef{RepoID: f.repo.ID, Name: "Deploy"}, Step: "test"}

		_, err := f.metrics.QuarantineStep(ctx, key(f), "", now)
		require.NoError(t, err)

		active, err := f.metrics.ActiveQuarantines(ctx, now)
		require.NoError(t, err)
		require.Contains(t, active, key(f))
		require.NotContains(t, active, other)
	})
}

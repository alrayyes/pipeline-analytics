package settings_test

import (
	"context"
	"maps"
	"sync"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/settings"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	mu   sync.Mutex
	data map[string]map[string]string // userID -> overrides
}

func newFakeStore() *fakeStore {
	return &fakeStore{data: map[string]map[string]string{}}
}

func (f *fakeStore) Get(_ context.Context, userID string) (map[string]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	overrides := map[string]string{}
	maps.Copy(overrides, f.data[userID])

	return overrides, nil
}

func (f *fakeStore) Patch(_ context.Context, userID string, updates map[string]*string) (map[string]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	overrides := f.data[userID]
	if overrides == nil {
		overrides = map[string]string{}
	}

	for k, v := range updates {
		if v == nil {
			delete(overrides, k)
		} else {
			overrides[k] = *v
		}
	}

	f.data[userID] = overrides
	result := map[string]string{}
	maps.Copy(result, overrides)

	return result, nil
}

func TestService_Get(t *testing.T) {
	t.Parallel()

	t.Run("returns documented defaults when nothing has ever been set", func(t *testing.T) {
		t.Parallel()

		svc := settings.NewService(newFakeStore())

		got, err := svc.Get(context.Background(), "user-1")
		require.NoError(t, err)
		require.Equal(t, settings.Settings{
			Theme:                 settings.DefaultTheme,
			ForgeFilter:           settings.DefaultForgeFilter,
			PipelinesHealthFilter: settings.DefaultPipelinesHealthFilter,
			PipelinesRepoSelector: settings.DefaultPipelinesRepoSelector,
			PipelinesSortOrder:    settings.DefaultPipelinesSortOrder,
		}, got)
	})

	t.Run("returns a stored override merged onto the remaining defaults", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		svc := settings.NewService(store)

		_, err := svc.Update(context.Background(), "user-1", map[string]*string{
			settings.KeyTheme: new("dark"),
		})
		require.NoError(t, err)

		got, err := svc.Get(context.Background(), "user-1")
		require.NoError(t, err)
		require.Equal(t, "dark", got.Theme)
		require.Equal(t, settings.DefaultForgeFilter, got.ForgeFilter)
	})
}

func TestService_Update(t *testing.T) {
	t.Parallel()

	t.Run("a value is visible on the next Get, from any caller", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		svc := settings.NewService(store)

		_, err := svc.Update(context.Background(), "user-1", map[string]*string{
			settings.KeyPipelinesSortOrder: new("lastRun"),
		})
		require.NoError(t, err)

		got, err := settings.NewService(store).Get(context.Background(), "user-1")
		require.NoError(t, err)
		require.Equal(t, "lastRun", got.PipelinesSortOrder)
	})

	t.Run("nil clears a key back to its default", func(t *testing.T) {
		t.Parallel()

		svc := settings.NewService(newFakeStore())
		ctx := context.Background()

		_, err := svc.Update(ctx, "user-1", map[string]*string{settings.KeyTheme: new("dark")})
		require.NoError(t, err)

		got, err := svc.Update(ctx, "user-1", map[string]*string{settings.KeyTheme: nil})
		require.NoError(t, err)
		require.Equal(t, settings.DefaultTheme, got.Theme)
	})

	t.Run("an invalid enum value is rejected and leaves the stored value unchanged", func(t *testing.T) {
		t.Parallel()

		svc := settings.NewService(newFakeStore())
		ctx := context.Background()

		_, err := svc.Update(ctx, "user-1", map[string]*string{settings.KeyTheme: new("dark")})
		require.NoError(t, err)

		_, err = svc.Update(ctx, "user-1", map[string]*string{settings.KeyTheme: new("not-a-real-theme")})
		require.ErrorIs(t, err, settings.ErrInvalidValue)

		got, err := svc.Get(ctx, "user-1")
		require.NoError(t, err)
		require.Equal(t, "dark", got.Theme, "unchanged by the rejected update")
	})

	t.Run("an unrecognized key is rejected", func(t *testing.T) {
		t.Parallel()

		svc := settings.NewService(newFakeStore())

		_, err := svc.Update(context.Background(), "user-1", map[string]*string{"notAKey": new("x")})
		require.ErrorIs(t, err, settings.ErrInvalidKey)
	})

	t.Run("a free-form value (PipelinesRepoSelector) isn't validated against an enum", func(t *testing.T) {
		t.Parallel()

		svc := settings.NewService(newFakeStore())

		got, err := svc.Update(context.Background(), "user-1", map[string]*string{
			settings.KeyPipelinesRepoSelector: new("some-repo-id"),
		})
		require.NoError(t, err)
		require.Equal(t, "some-repo-id", got.PipelinesRepoSelector)
	})

	t.Run("multiple keys in one request all apply together", func(t *testing.T) {
		t.Parallel()

		svc := settings.NewService(newFakeStore())

		got, err := svc.Update(context.Background(), "user-1", map[string]*string{
			settings.KeyTheme:       new("dark"),
			settings.KeyForgeFilter: new("github"),
		})
		require.NoError(t, err)
		require.Equal(t, "dark", got.Theme)
		require.Equal(t, "github", got.ForgeFilter)
	})

	t.Run("an invalid value in a multi-key request rejects the whole request", func(t *testing.T) {
		t.Parallel()

		svc := settings.NewService(newFakeStore())
		ctx := context.Background()

		_, err := svc.Update(ctx, "user-1", map[string]*string{
			settings.KeyTheme:       new("dark"),
			settings.KeyForgeFilter: new("not-a-real-forge"),
		})
		require.ErrorIs(t, err, settings.ErrInvalidValue)

		got, err := svc.Get(ctx, "user-1")
		require.NoError(t, err)
		require.Equal(t, settings.DefaultTheme, got.Theme, "neither key applied")
	})
}

func TestService_ResetPipelinesFilters(t *testing.T) {
	t.Parallel()

	svc := settings.NewService(newFakeStore())
	ctx := context.Background()

	_, err := svc.Update(ctx, "user-1", map[string]*string{
		settings.KeyPipelinesHealthFilter: new("healthy"),
		settings.KeyPipelinesRepoSelector: new("some-repo"),
		settings.KeyPipelinesSortOrder:    new("lastRun"),
		settings.KeyTheme:                 new("dark"),
	})
	require.NoError(t, err)

	got, err := svc.ResetPipelinesFilters(ctx, "user-1")
	require.NoError(t, err)
	require.Equal(t, settings.DefaultPipelinesHealthFilter, got.PipelinesHealthFilter)
	require.Equal(t, settings.DefaultPipelinesRepoSelector, got.PipelinesRepoSelector)
	require.Equal(t, settings.DefaultPipelinesSortOrder, got.PipelinesSortOrder)
	require.Equal(t, "dark", got.Theme, "not a Pipelines filter -- untouched by reset")
}

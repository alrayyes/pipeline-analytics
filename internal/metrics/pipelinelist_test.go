package metrics_test

import (
	"context"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

// healthyRuns and failingRuns give a pipeline a most recent run started at
// startMin, with a healthy or an unhealthy recent history.
func healthyRuns(startMin int) []metrics.RunRecord {
	return []metrics.RunRecord{
		completedRun(startMin, 5, "success"),
		completedRun(startMin-10, 5, "success"),
		completedRun(startMin-20, 5, "success"),
	}
}

func failingRuns(startMin int) []metrics.RunRecord {
	return []metrics.RunRecord{
		completedRun(startMin, 5, "failure"),
		completedRun(startMin-10, 5, "failure"),
		completedRun(startMin-20, 5, "success"),
	}
}

// fivePipelines: a, c and e healthy; b and d unhealthy. Last run, newest
// first: d (500), a (400), e (300), b (200), c (100), and "f" has no runs.
func fivePipelines() *fakeStore {
	refs := map[string]metrics.PipelineRef{}
	for _, name := range []string{"a", "b", "c", "d", "e", "f"} {
		refs[name] = metrics.PipelineRef{RepoID: "repo-1", Name: name}
	}

	return &fakeStore{
		pipelines: []metrics.PipelineRef{refs["a"], refs["b"], refs["c"], refs["d"], refs["e"], refs["f"]},
		runs: map[metrics.PipelineRef][]metrics.RunRecord{
			refs["a"]: healthyRuns(400),
			refs["b"]: failingRuns(200),
			refs["c"]: healthyRuns(100),
			refs["d"]: failingRuns(500),
			refs["e"]: healthyRuns(300),
		},
	}
}

func listNames(t *testing.T, filter metrics.PipelineListFilter) ([]string, bool) {
	t.Helper()

	pipelines, hasMore, err := metrics.NewService(fivePipelines()).ListPipelines(context.Background(), metrics.Window{RunCount: 10}, filter)
	require.NoError(t, err)

	names := make([]string, 0, len(pipelines))
	for _, p := range pipelines {
		names = append(names, p.Name)
	}

	return names, hasMore
}

func TestService_ListPipelines_HealthFilter(t *testing.T) {
	t.Parallel()

	t.Run("returns only the matching pipelines, from the whole set", func(t *testing.T) {
		t.Parallel()

		names, _ := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthUnhealthy})

		require.Equal(t, []string{"b", "d"}, names)
	})

	t.Run("a match on a later page is found, and every page but the last is full", func(t *testing.T) {
		t.Parallel()

		first, hasMore := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthUnhealthy, Limit: 1})
		require.Equal(t, []string{"b"}, first)
		require.True(t, hasMore)

		second, hasMore := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthUnhealthy, Limit: 1, Offset: 1})
		require.Equal(t, []string{"d"}, second)
		require.False(t, hasMore)
	})

	t.Run("healthy pages hold their full size over the filtered set", func(t *testing.T) {
		t.Parallel()

		// a, c, e and f are healthy or have no failing runs; the page counts
		// them, not the unhealthy pipelines between them.
		first, hasMore := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthHealthy, Limit: 2})
		require.Len(t, first, 2)
		require.True(t, hasMore)

		last, hasMore := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthHealthy, Limit: 2, Offset: 2})
		require.NotEmpty(t, last)
		require.False(t, hasMore)
	})

	t.Run("an offset past the matches is an empty page, not an error", func(t *testing.T) {
		t.Parallel()

		names, hasMore := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthUnhealthy, Limit: 5, Offset: 10})

		require.Empty(t, names)
		require.False(t, hasMore)
	})

	t.Run("composes with the repo filter", func(t *testing.T) {
		t.Parallel()

		names, _ := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthUnhealthy, RepoID: "repo-2"})

		require.Empty(t, names)
	})
}

func TestService_ListPipelines_Sort(t *testing.T) {
	t.Parallel()

	t.Run("lastRun puts the most recently run first and a pipeline with no runs last", func(t *testing.T) {
		t.Parallel()

		names, _ := listNames(t, metrics.PipelineListFilter{Sort: metrics.SortByLastRun})

		require.Equal(t, []string{"d", "a", "e", "b", "c", "f"}, names)
	})

	t.Run("name is the default order", func(t *testing.T) {
		t.Parallel()

		byDefault, _ := listNames(t, metrics.PipelineListFilter{})
		byName, _ := listNames(t, metrics.PipelineListFilter{Sort: metrics.SortByName})

		require.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, byDefault)
		require.Equal(t, byDefault, byName)
	})

	t.Run("the sort applies across pages, not within one", func(t *testing.T) {
		t.Parallel()

		first, hasMore := listNames(t, metrics.PipelineListFilter{Sort: metrics.SortByLastRun, Limit: 2})
		require.Equal(t, []string{"d", "a"}, first)
		require.True(t, hasMore)

		second, _ := listNames(t, metrics.PipelineListFilter{Sort: metrics.SortByLastRun, Limit: 2, Offset: 2})
		require.Equal(t, []string{"e", "b"}, second)
	})

	t.Run("with a health filter, the matches are sorted then paged", func(t *testing.T) {
		t.Parallel()

		names, hasMore := listNames(t, metrics.PipelineListFilter{Health: metrics.HealthHealthy, Sort: metrics.SortByLastRun, Limit: 2})

		require.Equal(t, []string{"a", "e"}, names)
		require.True(t, hasMore)
	})
}

func TestParseHealthFilter(t *testing.T) {
	t.Parallel()

	for raw, want := range map[string]metrics.HealthStatus{
		"":          "",
		"healthy":   metrics.HealthHealthy,
		"unhealthy": metrics.HealthUnhealthy,
	} {
		got, err := metrics.ParseHealthFilter(raw)
		require.NoError(t, err, raw)
		require.Equal(t, want, got, raw)
	}

	for _, raw := range []string{"all", "Healthy", "broken", "1"} {
		_, err := metrics.ParseHealthFilter(raw)
		require.ErrorIs(t, err, metrics.ErrInvalidHealthFilter, raw)
	}
}

func TestParsePipelineSort(t *testing.T) {
	t.Parallel()

	for raw, want := range map[string]metrics.PipelineSort{
		"":        metrics.SortByName,
		"name":    metrics.SortByName,
		"lastRun": metrics.SortByLastRun,
	} {
		got, err := metrics.ParsePipelineSort(raw)
		require.NoError(t, err, raw)
		require.Equal(t, want, got, raw)
	}

	for _, raw := range []string{"lastrun", "last_run", "newest", "1"} {
		_, err := metrics.ParsePipelineSort(raw)
		require.ErrorIs(t, err, metrics.ErrInvalidPipelineSort, raw)
	}
}

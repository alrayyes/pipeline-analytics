package sqlite_test

import (
	"context"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestStore_WindowBranches(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	for i, run := range []ingestion.Run{
		{ForgeRunID: "1", Branch: "main"},
		{ForgeRunID: "2", Branch: "main"},
		{ForgeRunID: "3", Branch: "feature/b"},
		{ForgeRunID: "4", Branch: "feature/a"},
		{ForgeRunID: "5", Branch: ""},
		{ForgeRunID: "6", Branch: "outside"},
	} {
		run.Status = "completed"
		run.Conclusion = "success"
		run.StartedAt = at(100 + i)

		if run.ForgeRunID == "6" {
			run.StartedAt = at(900)
		}

		f.seedRunWith(t, run)
	}

	window := metrics.RunWindowFilter{Since: *at(50), Until: *at(500)}

	t.Run("counts runs per branch in the window, busiest first then by name, skipping runs with no branch", func(t *testing.T) {
		t.Parallel()

		got, err := f.metrics.WindowBranches(context.Background(), window)
		require.NoError(t, err)
		require.Equal(t, []metrics.BranchCount{
			{Name: "main", RunCount: 2},
			{Name: "feature/a", RunCount: 1},
			{Name: "feature/b", RunCount: 1},
		}, got)
	})

	t.Run("scopes to a repo and a forge, and ignores a branch filter", func(t *testing.T) {
		t.Parallel()

		scoped := window
		scoped.RepoID = "nope"

		got, err := f.metrics.WindowBranches(context.Background(), scoped)
		require.NoError(t, err)
		require.Empty(t, got)

		scoped = window
		scoped.Forge = "forgejo"

		got, err = f.metrics.WindowBranches(context.Background(), scoped)
		require.NoError(t, err)
		require.Empty(t, got)

		scoped = window
		scoped.Forge = "github"
		scoped.Branch = "main"

		got, err = f.metrics.WindowBranches(context.Background(), scoped)
		require.NoError(t, err)
		require.Len(t, got, 3, "the branch list isn't narrowed by a branch")
	})
}

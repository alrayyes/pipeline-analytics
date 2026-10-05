package metrics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestService_ListBranches(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{windowBranches: []metrics.BranchCount{{Name: "main", RunCount: 3}}}

	got, err := metrics.NewService(store).ListBranches(
		context.Background(), now, metrics.ParseInsightWindow("24h"), metrics.InsightFilter{RepoID: "r1", Forge: "github"},
	)
	require.NoError(t, err)
	require.Equal(t, []metrics.BranchCount{{Name: "main", RunCount: 3}}, got)

	require.Equal(t, metrics.RunWindowFilter{
		RepoID: "r1",
		Forge:  "github",
		Since:  now.Add(-24 * time.Hour),
		Until:  now,
	}, store.gotBranchFilter, "the window is the trailing span ending now, and the filter's repo and forge pass through")
}

var errBranchesStore = errors.New("store down")

func TestService_ListBranches_StoreError(t *testing.T) {
	t.Parallel()

	_, err := metrics.NewService(&fakeStore{branchesErr: errBranchesStore}).ListBranches(
		context.Background(), time.Now(), metrics.ParseInsightWindow(""), metrics.InsightFilter{},
	)
	require.ErrorIs(t, err, errBranchesStore)
}

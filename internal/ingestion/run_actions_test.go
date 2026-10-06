package ingestion_test

import (
	"context"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

type fakeRunLookup struct {
	run ingestion.Run
	err error
}

func (f fakeRunLookup) RunByID(context.Context, string) (ingestion.Run, error) {
	return f.run, f.err
}

type fakeActor struct {
	err error
	got ingestion.RunActionRequest
	did string
}

func (f *fakeActor) RerunRun(_ context.Context, req ingestion.RunActionRequest) error {
	f.got, f.did = req, "rerun"

	return f.err
}

func (f *fakeActor) CancelRun(_ context.Context, req ingestion.RunActionRequest) error {
	f.got, f.did = req, "cancel"

	return f.err
}

func newActionService(run ingestion.Run, runErr error, repos fakeLogRepos, actor *fakeActor) *ingestion.RunActionService {
	actors := map[ingestion.Forge]ingestion.RunActor{}
	if actor != nil {
		actors[ingestion.ForgeGitHub] = actor
	}

	return ingestion.NewRunActionService(fakeRunLookup{run: run, err: runErr}, repos, actors)
}

func githubRepos() fakeLogRepos {
	return fakeLogRepos{repo: ingestion.Repo{ID: "r1", Forge: ingestion.ForgeGitHub, Identifier: "o/n"}}
}

func TestRunActionService_Rerun(t *testing.T) {
	t.Parallel()

	failed := ingestion.Run{RepoID: "r1", ForgeRunID: "9", Status: "completed", Conclusion: "failure"}

	t.Run("asks the forge to re-run the failed jobs of a failed run", func(t *testing.T) {
		t.Parallel()

		actor := &fakeActor{}

		require.NoError(t, newActionService(failed, nil, githubRepos(), actor).Rerun(context.Background(), "run1"))
		require.True(t, actor.got.FailedOnly)
	})

	t.Run("a timed-out run counts as failed", func(t *testing.T) {
		t.Parallel()

		actor := &fakeActor{}
		run := ingestion.Run{RepoID: "r1", Status: "completed", Conclusion: "timed_out"}

		require.NoError(t, newActionService(run, nil, githubRepos(), actor).Rerun(context.Background(), "run1"))
		require.True(t, actor.got.FailedOnly)
	})

	t.Run("a run still in progress is not actionable", func(t *testing.T) {
		t.Parallel()

		run := ingestion.Run{RepoID: "r1", Status: "in_progress"}

		err := newActionService(run, nil, githubRepos(), &fakeActor{}).Rerun(context.Background(), "run1")

		require.ErrorIs(t, err, ingestion.ErrNotActionable)
	})

	t.Run("an unknown run is reported as such", func(t *testing.T) {
		t.Parallel()

		err := newActionService(ingestion.Run{}, ingestion.ErrRunNotFound, githubRepos(), &fakeActor{}).Rerun(context.Background(), "x")

		require.ErrorIs(t, err, ingestion.ErrRunNotFound)
	})

	t.Run("a forge with no actor is unsupported", func(t *testing.T) {
		t.Parallel()

		err := newActionService(failed, nil, githubRepos(), nil).Rerun(context.Background(), "run1")

		require.ErrorIs(t, err, ingestion.ErrActionUnsupported)
	})

	t.Run("a refusal from the forge comes through", func(t *testing.T) {
		t.Parallel()

		err := newActionService(failed, nil, githubRepos(), &fakeActor{err: ingestion.ErrActionForbidden}).Rerun(context.Background(), "run1")

		require.ErrorIs(t, err, ingestion.ErrActionForbidden)
	})

	t.Run("a repo lookup failure is an error", func(t *testing.T) {
		t.Parallel()

		repos := githubRepos()
		repos.repoErr = errFakeBoom

		err := newActionService(failed, nil, repos, &fakeActor{}).Rerun(context.Background(), "run1")

		require.ErrorIs(t, err, errFakeBoom)
	})

	t.Run("a token lookup failure is an error", func(t *testing.T) {
		t.Parallel()

		repos := githubRepos()
		repos.tokenErr = errFakeBoom

		err := newActionService(failed, nil, repos, &fakeActor{}).Rerun(context.Background(), "run1")

		require.ErrorIs(t, err, errFakeBoom)
	})
}

func TestRunActionService_Cancel(t *testing.T) {
	t.Parallel()

	t.Run("cancels a run that hasn't concluded", func(t *testing.T) {
		t.Parallel()

		actor := &fakeActor{}
		run := ingestion.Run{RepoID: "r1", Status: "queued"}

		require.NoError(t, newActionService(run, nil, githubRepos(), actor).Cancel(context.Background(), "run1"))
		require.Equal(t, "cancel", actor.did)
	})

	t.Run("a concluded run is not actionable", func(t *testing.T) {
		t.Parallel()

		run := ingestion.Run{RepoID: "r1", Status: "completed", Conclusion: "success"}

		err := newActionService(run, nil, githubRepos(), &fakeActor{}).Cancel(context.Background(), "run1")

		require.ErrorIs(t, err, ingestion.ErrNotActionable)
	})
}

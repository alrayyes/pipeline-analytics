package ingestion_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

type fakeLocator struct {
	location ingestion.JobLocation
	err      error
}

func (f fakeLocator) LocateJob(context.Context, string, string) (ingestion.JobLocation, error) {
	return f.location, f.err
}

var errFakeBoom = errors.New("boom")

type fakeLogRepos struct {
	repo     ingestion.Repo
	repoErr  error
	tokenErr error
}

func (f fakeLogRepos) GetRepo(context.Context, string) (ingestion.Repo, error) {
	return f.repo, f.repoErr
}

func (f fakeLogRepos) RepoToken(context.Context, string) (string, error) {
	return "secret", f.tokenErr
}

type fakeReader struct {
	tail ingestion.JobLogTail
	err  error
	got  ingestion.JobLogRequest
}

func (f *fakeReader) JobLogTail(_ context.Context, req ingestion.JobLogRequest) (ingestion.JobLogTail, error) {
	f.got = req

	return f.tail, f.err
}

func newLogService(reader *fakeReader, forge ingestion.Forge) *ingestion.JobLogService {
	readers := map[ingestion.Forge]ingestion.JobLogReader{}
	if reader != nil {
		readers[forge] = reader
	}

	return ingestion.NewJobLogService(
		fakeLocator{location: ingestion.JobLocation{RepoID: "r1", ForgeJobID: "5001", ForgeURL: "https://forge/job/5001"}},
		fakeLogRepos{repo: ingestion.Repo{ID: "r1", Forge: forge, Identifier: "o/n", ForgejoInstanceURL: "https://forgejo.example"}},
		readers,
	)
}

func TestJobLogService_Tail(t *testing.T) {
	t.Parallel()

	t.Run("hands the reader the repo, token and forge job id, and returns its lines", func(t *testing.T) {
		t.Parallel()

		reader := &fakeReader{tail: ingestion.JobLogTail{Lines: []string{"a", "b"}, Truncated: true}}

		got, err := newLogService(reader, ingestion.ForgeGitHub).Tail(context.Background(), "run", "job", 50)
		require.NoError(t, err)

		require.True(t, got.Available)
		require.Equal(t, []string{"a", "b"}, got.Lines)
		require.True(t, got.Truncated)
		require.Equal(t, "https://forge/job/5001", got.ForgeURL)
		require.Equal(t, ingestion.JobLogRequest{
			InstanceURL: "https://forgejo.example", Identifier: "o/n", Token: "secret", ForgeJobID: "5001", Lines: 50,
		}, reader.got)
	})

	t.Run("no lines asked for means 200, and the most is 1000", func(t *testing.T) {
		t.Parallel()

		reader := &fakeReader{}
		svc := newLogService(reader, ingestion.ForgeGitHub)

		_, err := svc.Tail(context.Background(), "run", "job", 0)
		require.NoError(t, err)
		require.Equal(t, ingestion.DefaultLogLines, reader.got.Lines)

		_, err = svc.Tail(context.Background(), "run", "job", 99999)
		require.NoError(t, err)
		require.Equal(t, ingestion.MaxLogLines, reader.got.Lines)
	})

	for _, tc := range []struct {
		name string
		err  error
		want ingestion.JobLogReason
	}{
		{"unsupported", ingestion.ErrLogUnsupported, ingestion.LogUnsupported},
		{"expired", ingestion.ErrLogExpired, ingestion.LogExpired},
		{"forbidden", ingestion.ErrLogForbidden, ingestion.LogForbidden},
		{"unreachable", ingestion.ErrLogUnreachable, ingestion.LogUnreachable},
	} {
		t.Run("a "+tc.name+" log is a result with a reason and the deep link, not an error", func(t *testing.T) {
			t.Parallel()

			got, err := newLogService(&fakeReader{err: tc.err}, ingestion.ForgeGitHub).Tail(context.Background(), "run", "job", 10)
			require.NoError(t, err)
			require.False(t, got.Available)
			require.Equal(t, tc.want, got.Reason)
			require.Empty(t, got.Lines)
			require.Equal(t, "https://forge/job/5001", got.ForgeURL)
		})
	}

	t.Run("a forge with no reader is unsupported", func(t *testing.T) {
		t.Parallel()

		got, err := newLogService(nil, ingestion.ForgeForgejo).Tail(context.Background(), "run", "job", 10)
		require.NoError(t, err)
		require.Equal(t, ingestion.LogUnsupported, got.Reason)
	})

	t.Run("an unknown job is an error", func(t *testing.T) {
		t.Parallel()

		svc := ingestion.NewJobLogService(fakeLocator{err: ingestion.ErrJobNotFound}, fakeLogRepos{}, nil)

		_, err := svc.Tail(context.Background(), "run", "nope", 10)
		require.ErrorIs(t, err, ingestion.ErrJobNotFound)
	})

	t.Run("any other reader failure is an error, not a reason", func(t *testing.T) {
		t.Parallel()

		_, err := newLogService(&fakeReader{err: errFakeBoom}, ingestion.ForgeGitHub).Tail(context.Background(), "run", "job", 10)
		require.ErrorIs(t, err, errFakeBoom)
	})

	t.Run("a repo or token lookup failure is an error", func(t *testing.T) {
		t.Parallel()

		located := fakeLocator{location: ingestion.JobLocation{RepoID: "r1"}}
		readers := map[ingestion.Forge]ingestion.JobLogReader{ingestion.ForgeGitHub: &fakeReader{}}

		_, err := ingestion.NewJobLogService(located, fakeLogRepos{repoErr: errFakeBoom}, readers).Tail(context.Background(), "run", "job", 10)
		require.ErrorIs(t, err, errFakeBoom)

		repo := ingestion.Repo{ID: "r1", Forge: ingestion.ForgeGitHub}

		_, err = ingestion.NewJobLogService(located, fakeLogRepos{repo: repo, tokenErr: errFakeBoom}, readers).Tail(context.Background(), "run", "job", 10)
		require.ErrorIs(t, err, errFakeBoom)
	})
}

package ingestion_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

var errInsufficientScope = errors.New("insufficient scope")

type fakeStore struct {
	mu    sync.Mutex
	repos map[string]ingestion.Repo
	seq   int
}

func newFakeStore() *fakeStore {
	return &fakeStore{repos: map[string]ingestion.Repo{}}
}

func (f *fakeStore) CreateRepo(_ context.Context, in ingestion.NewRepo) (ingestion.Repo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.seq++
	repo := ingestion.Repo{
		ID:                 fmt.Sprintf("repo-%d", f.seq),
		Forge:              in.Forge,
		Identifier:         in.Identifier,
		ForgejoInstanceURL: in.ForgejoInstanceURL,
		TokenMasked:        ingestion.MaskToken(in.Token),
		WebhookSecret:      "secret",
		IngestionStatus:    ingestion.StatusPending,
	}
	f.repos[repo.ID] = repo

	return repo, nil
}

func (f *fakeStore) ListRepos(context.Context) ([]ingestion.Repo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	repos := make([]ingestion.Repo, 0, len(f.repos))
	for _, r := range f.repos {
		repos = append(repos, r)
	}

	return repos, nil
}

var errRepoNotFound = errors.New("repo not found")

func (f *fakeStore) GetRepo(_ context.Context, id string) (ingestion.Repo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	repo, ok := f.repos[id]
	if !ok {
		return ingestion.Repo{}, errRepoNotFound
	}

	return repo, nil
}

func (f *fakeStore) RepoToken(context.Context, string) (string, error) {
	return "", nil
}

func (f *fakeStore) SetIngestionStatus(_ context.Context, id string, status ingestion.Status, reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	repo, ok := f.repos[id]
	if !ok {
		return errRepoNotFound
	}

	repo.IngestionStatus = status
	repo.IngestionStatusReason = reason
	f.repos[id] = repo

	return nil
}

func (f *fakeStore) SetReconcileETag(_ context.Context, id string, etag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	repo, ok := f.repos[id]
	if !ok {
		return errRepoNotFound
	}

	repo.ReconcileETag = etag
	f.repos[id] = repo

	return nil
}

type fakeForgeClient struct {
	err          error
	listRunsFunc func(context.Context, ingestion.ListRunsRequest) (ingestion.ListRunsResult, error)
}

func (f *fakeForgeClient) CreateWebhook(context.Context, ingestion.CreateWebhookRequest) error {
	return f.err
}

func (f *fakeForgeClient) ListRecentRuns(ctx context.Context, req ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
	if f.listRunsFunc != nil {
		return f.listRunsFunc(ctx, req)
	}

	return ingestion.ListRunsResult{}, nil
}

func TestRegistrar_Register(t *testing.T) {
	t.Parallel()

	t.Run("successful webhook creation marks the repo active", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{},
		}, "https://example.com")

		repo, err := registrar.Register(context.Background(), ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
		})
		require.NoError(t, err)
		require.Equal(t, ingestion.StatusActive, repo.IngestionStatus)
	})

	t.Run("webhook creation failure degrades the repo instead of failing registration", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{err: errInsufficientScope},
		}, "https://example.com")

		repo, err := registrar.Register(context.Background(), ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
		})
		require.NoError(t, err)
		require.Equal(t, ingestion.StatusDegraded, repo.IngestionStatus)
		require.Contains(t, repo.IngestionStatusReason, "insufficient scope")
	})

	t.Run("no client configured for the forge degrades the repo", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{}, "https://example.com")

		repo, err := registrar.Register(context.Background(), ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
		})
		require.NoError(t, err)
		require.Equal(t, ingestion.StatusDegraded, repo.IngestionStatus)
	})
}

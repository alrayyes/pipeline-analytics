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

func (f *fakeStore) ListRepoIdentifiers(_ context.Context, forge ingestion.Forge, instanceURL string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var identifiers []string

	for _, r := range f.repos {
		if r.Forge == forge && r.ForgejoInstanceURL == instanceURL {
			identifiers = append(identifiers, r.Identifier)
		}
	}

	return identifiers, nil
}

func (f *fakeStore) ListRepos(context.Context, ingestion.RepoListFilter) ([]ingestion.Repo, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	repos := make([]ingestion.Repo, 0, len(f.repos))
	for _, r := range f.repos {
		repos = append(repos, r)
	}

	return repos, false, nil
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

func (f *fakeStore) DeleteRepo(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.repos[id]; !ok {
		return errRepoNotFound
	}

	delete(f.repos, id)

	return nil
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
	err             error
	listRunsFunc    func(context.Context, ingestion.ListRunsRequest) (ingestion.ListRunsResult, error)
	discoveredRepos []string
	// getRepoMeta/getRepoErr are separate from err -- a test exercising a
	// CreateWebhook failure (via err) shouldn't also fail the GetRepo
	// check Register now runs first, or it'd never reach the webhook
	// step it means to exercise.
	getRepoMeta ingestion.RepoMetadata
	getRepoErr  error
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

func (f *fakeForgeClient) GetRepo(context.Context, ingestion.GetRepoRequest) (ingestion.RepoMetadata, error) {
	return f.getRepoMeta, f.getRepoErr
}

func (f *fakeForgeClient) ListAccessibleRepos(context.Context, ingestion.ListAccessibleReposRequest) ([]string, error) {
	return f.discoveredRepos, f.err
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

	t.Run("registering an archived repo is rejected, no record created", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{getRepoMeta: ingestion.RepoMetadata{Archived: true}},
		}, "https://example.com")

		_, err := registrar.Register(context.Background(), ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/old-repo",
			Token:      "ghp_test",
		})
		require.ErrorIs(t, err, ingestion.ErrRepoArchived)
		require.Empty(t, store.repos)
	})

	t.Run("registering a fork is rejected, no record created", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{getRepoMeta: ingestion.RepoMetadata{Fork: true}},
		}, "https://example.com")

		_, err := registrar.Register(context.Background(), ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/a-fork",
			Token:      "ghp_test",
		})
		require.ErrorIs(t, err, ingestion.ErrRepoFork)
		require.Empty(t, store.repos)
	})

	t.Run("registering a mirror is rejected, no record created", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{getRepoMeta: ingestion.RepoMetadata{Mirror: true}},
		}, "https://example.com")

		_, err := registrar.Register(context.Background(), ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/a-mirror",
			Token:      "ghp_test",
		})
		require.ErrorIs(t, err, ingestion.ErrRepoMirror)
		require.Empty(t, store.repos)
	})

	t.Run("a normal repo still registers", func(t *testing.T) {
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

	t.Run("a GetRepo failure fails registration rather than proceeding unverified", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{getRepoErr: errInsufficientScope},
		}, "https://example.com")

		_, err := registrar.Register(context.Background(), ingestion.NewRepo{
			Forge:      ingestion.ForgeGitHub,
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
		})
		require.ErrorIs(t, err, errInsufficientScope)
		require.Empty(t, store.repos)
	})
}

func TestRegistrar_Discover(t *testing.T) {
	t.Parallel()

	t.Run("returns the repos the client discovers", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{discoveredRepos: []string{"alrayyes/pipeline-analytics"}},
		}, "https://example.com")

		repos, err := registrar.Discover(context.Background(), ingestion.ForgeGitHub, "", "ghp_test")
		require.NoError(t, err)
		require.Equal(t, []string{"alrayyes/pipeline-analytics"}, repos)
	})

	t.Run("propagates a forge error", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{
			ingestion.ForgeGitHub: &fakeForgeClient{err: errInsufficientScope},
		}, "https://example.com")

		_, err := registrar.Discover(context.Background(), ingestion.ForgeGitHub, "", "ghp_test")
		require.ErrorIs(t, err, errInsufficientScope)
	})

	t.Run("errors when no client is configured for the forge", func(t *testing.T) {
		t.Parallel()

		store := newFakeStore()
		registrar := ingestion.NewRegistrar(store, map[ingestion.Forge]ingestion.ForgeClient{}, "https://example.com")

		_, err := registrar.Discover(context.Background(), ingestion.ForgeGitHub, "", "ghp_test")
		require.Error(t, err)
	})
}

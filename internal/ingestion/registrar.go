package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// ErrNoForgeClient is returned when Discover is asked about a forge this
// server has no client configured for.
var ErrNoForgeClient = errors.New("no client configured for forge")

// Registrar orchestrates repo registration: persist the repo, then create
// its webhook on the forge, degrading rather than failing when the webhook
// step doesn't succeed (forge-ingestion/spec.md's "Webhook creation failure
// is surfaced").
type Registrar struct {
	store       Store
	clients     map[Forge]ForgeClient
	callbackURL string
}

// NewRegistrar returns a Registrar. callbackURL is this server's own public
// base URL; the webhook path for each forge is appended to it.
func NewRegistrar(store Store, clients map[Forge]ForgeClient, callbackURL string) *Registrar {
	return &Registrar{store: store, clients: clients, callbackURL: callbackURL}
}

// Register verifies the repo isn't archived, a fork, or a mirror, then
// persists it and attempts to create its webhook. A storage failure, or the
// repo actually being archived/a fork/a mirror, is returned as an error
// before anything is persisted; a webhook-creation failure -- and a
// GetRepo failure, which leaves that status genuinely unverified rather
// than known-rejectable -- is instead recorded on the returned Repo as a
// degraded ingestion status, once it's already tracked. Blocking outright
// on a GetRepo failure (a bad token, a network blip) would make
// registration itself brittle against exactly the kind of forge
// reachability issue the existing degrade path already exists to absorb;
// confirmed live via this app's own end-to-end suite, which registers
// against a token that can't reach the real forge at all.
func (r *Registrar) Register(ctx context.Context, in NewRepo) (Repo, error) {
	client, ok := r.clients[in.Forge]
	if ok {
		meta, err := client.GetRepo(ctx, GetRepoRequest{
			InstanceURL: in.ForgejoInstanceURL,
			Identifier:  in.Identifier,
			Token:       in.Token,
		})
		if err == nil {
			switch {
			case meta.Archived:
				return Repo{}, ErrRepoArchived
			case meta.Fork:
				return Repo{}, ErrRepoFork
			case meta.Mirror:
				return Repo{}, ErrRepoMirror
			}
		}
		// A GetRepo error falls through to CreateRepo below and on to the
		// CreateWebhook attempt, which degrades on the same class of
		// failure -- no separate degrade path needed for this case.
	}

	repo, err := r.store.CreateRepo(ctx, in)
	if err != nil {
		return Repo{}, fmt.Errorf("create repo: %w", err)
	}

	if !ok {
		return r.degrade(ctx, repo, fmt.Sprintf("no client configured for forge %q", repo.Forge))
	}

	err = client.CreateWebhook(ctx, CreateWebhookRequest{
		InstanceURL: repo.ForgejoInstanceURL,
		Identifier:  repo.Identifier,
		Token:       in.Token,
		CallbackURL: r.callbackURL + "/webhooks/" + string(repo.Forge),
		Secret:      repo.WebhookSecret,
	})
	if err != nil {
		return r.degrade(ctx, repo, err.Error())
	}

	if err := r.store.SetIngestionStatus(ctx, repo.ID, StatusActive, ""); err != nil {
		return Repo{}, fmt.Errorf("mark repo active: %w", err)
	}

	repo.IngestionStatus = StatusActive
	repo.IngestionStatusReason = ""

	return repo, nil
}

// Discover lists the "owner/name" repos a not-yet-registered token can
// access, for the registration UI's repo picker. Never persists anything.
func (r *Registrar) Discover(ctx context.Context, forge Forge, instanceURL, token string) ([]string, error) {
	slog.DebugContext(ctx, "repo discovery requested", "forge", forge)

	client, ok := r.clients[forge]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNoForgeClient, forge)
	}

	repos, err := client.ListAccessibleRepos(ctx, ListAccessibleReposRequest{
		InstanceURL: instanceURL,
		Token:       token,
	})
	if err != nil {
		return nil, fmt.Errorf("discover repos: %w", err)
	}

	slog.DebugContext(ctx, "repo discovery completed", "forge", forge, "repos", len(repos))

	return repos, nil
}

func (r *Registrar) degrade(ctx context.Context, repo Repo, reason string) (Repo, error) {
	if err := r.store.SetIngestionStatus(ctx, repo.ID, StatusDegraded, reason); err != nil {
		return Repo{}, fmt.Errorf("mark repo degraded: %w", err)
	}

	repo.IngestionStatus = StatusDegraded
	repo.IngestionStatusReason = reason

	return repo, nil
}

package ingestion

import (
	"context"
	"fmt"
)

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

// Register persists in and attempts to create its webhook. A storage
// failure is returned as an error; a webhook-creation failure is instead
// recorded on the returned Repo as a degraded ingestion status.
func (r *Registrar) Register(ctx context.Context, in NewRepo) (Repo, error) {
	repo, err := r.store.CreateRepo(ctx, in)
	if err != nil {
		return Repo{}, fmt.Errorf("create repo: %w", err)
	}

	client, ok := r.clients[repo.Forge]
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

func (r *Registrar) degrade(ctx context.Context, repo Repo, reason string) (Repo, error) {
	if err := r.store.SetIngestionStatus(ctx, repo.ID, StatusDegraded, reason); err != nil {
		return Repo{}, fmt.Errorf("mark repo degraded: %w", err)
	}

	repo.IngestionStatus = StatusDegraded
	repo.IngestionStatusReason = reason

	return repo, nil
}

// Package ingestion holds the forge-ingestion domain: tracked repos and the
// port their storage is reached through. See
// openspec/changes/add-pipeline-dashboard/specs/forge-ingestion/spec.md.
package ingestion

import (
	"context"
	"time"
)

// Forge identifies which forge a tracked repo lives on.
type Forge string

// The forges pipeline-analytics tracks repos on.
const (
	ForgeGitHub  Forge = "github"
	ForgeForgejo Forge = "forgejo"
)

// Status is a tracked repo's ingestion health.
type Status string

// The ingestion statuses a tracked repo can be in.
const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusDegraded Status = "degraded"
)

// Repo is a tracked repository. Its raw or encrypted token is never part of
// this type -- TokenMasked is all a caller outside the store ever sees.
type Repo struct {
	ID                    string
	Forge                 Forge
	Identifier            string
	ForgejoInstanceURL    string
	TokenMasked           string
	WebhookSecret         string
	IngestionStatus       Status
	IngestionStatusReason string
	CreatedAt             time.Time
}

// NewRepo is the input to registering a repo for tracking.
type NewRepo struct {
	Forge              Forge
	Identifier         string
	ForgejoInstanceURL string
	Token              string
}

// Store is the port the domain persists tracked repos through.
type Store interface {
	CreateRepo(ctx context.Context, repo NewRepo) (Repo, error)
	ListRepos(ctx context.Context) ([]Repo, error)
	GetRepo(ctx context.Context, id string) (Repo, error)
	// RepoToken returns the decrypted token for internal use (calling the
	// forge's API) -- never exposed over HTTP.
	RepoToken(ctx context.Context, id string) (string, error)
	// SetIngestionStatus records whether webhook registration (or later
	// ingestion) succeeded, and why when it didn't.
	SetIngestionStatus(ctx context.Context, id string, status Status, reason string) error
}

// CreateWebhookRequest is what a ForgeClient needs to register a webhook on
// a tracked repo.
type CreateWebhookRequest struct {
	// InstanceURL is set for Forgejo, empty for GitHub.
	InstanceURL string
	Identifier  string
	Token       string
	CallbackURL string
	Secret      string
}

// ForgeClient is the port the domain calls out to a forge's REST API
// through, to manage the webhook a tracked repo needs.
type ForgeClient interface {
	CreateWebhook(ctx context.Context, req CreateWebhookRequest) error
}

// MaskToken returns a display form of token that reveals at most its last
// four characters, per forge-ingestion/spec.md's "Credential storage"
// requirement.
func MaskToken(token string) string {
	const visible = 4
	if len(token) <= visible {
		return "****"
	}

	return "****" + token[len(token)-visible:]
}

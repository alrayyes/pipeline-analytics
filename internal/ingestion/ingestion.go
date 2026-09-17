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
	// ReconcileETag is the response ETag from the repo's last reconciliation
	// poll (GitHub only), used as If-None-Match on the next poll so an
	// unchanged repo costs no rate-limit budget. Empty before the first poll.
	ReconcileETag string
	CreatedAt     time.Time
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
	// DeleteRepo stops tracking a repo. The schema's ON DELETE CASCADE
	// removes its stored runs, jobs, and steps along with it -- it does
	// not attempt to delete the webhook created on the forge, since no
	// webhook ID is stored at registration time to delete it by.
	DeleteRepo(ctx context.Context, id string) error
	// RepoToken returns the decrypted token for internal use (calling the
	// forge's API) -- never exposed over HTTP.
	RepoToken(ctx context.Context, id string) (string, error)
	// SetIngestionStatus records whether webhook registration (or later
	// ingestion) succeeded, and why when it didn't.
	SetIngestionStatus(ctx context.Context, id string, status Status, reason string) error
	// SetReconcileETag records the ETag from a repo's most recent
	// reconciliation poll, for use as If-None-Match on the next one.
	SetReconcileETag(ctx context.Context, id string, etag string) error
}

// RepoReconciler is the port webhook ingestion calls to trigger an
// immediate reconciliation poll for one repo, for a forge whose webhook
// delivery carries no usable run data of its own -- Forgejo, see
// ProcessForgejoEvent.
type RepoReconciler interface {
	ReconcileRepo(ctx context.Context, repo Repo) error
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
// through, to manage the webhook a tracked repo needs and to reconcile
// tracked repos' recent run history.
type ForgeClient interface {
	CreateWebhook(ctx context.Context, req CreateWebhookRequest) error
	// ListRecentRuns lists a tracked repo's recent workflow runs (with their
	// jobs and steps) for reconciliation polling. On GitHub, req.ETag is
	// sent as If-None-Match; ListRunsResult.NotModified is true when the
	// forge reports nothing changed, costing no rate-limit budget.
	ListRecentRuns(ctx context.Context, req ListRunsRequest) (ListRunsResult, error)
	// ListAccessibleRepos lists "owner/name" identifiers the given token can
	// access, for the registration UI's repo picker. Called with a token
	// that hasn't been stored yet -- never persisted or logged.
	ListAccessibleRepos(ctx context.Context, req ListAccessibleReposRequest) ([]string, error)
}

// RateLimitSnapshot is a forge API's most recently observed rate-limit
// status for one token.
type RateLimitSnapshot struct {
	Limit     int
	Remaining int
	Used      int
	Resource  string
	ResetAt   time.Time
}

// RateLimitReporter is implemented by a ForgeClient that can report the
// rate-limit status it last observed a token being given, read from a real
// response's headers rather than a dedicated poll -- GitHub only; Forgejo
// has no equivalent concept and its ForgeClient doesn't implement this.
type RateLimitReporter interface {
	// RateLimitFor returns the token's most recently observed status. ok is
	// false if this token hasn't been used on a request yet.
	RateLimitFor(token string) (snapshot RateLimitSnapshot, ok bool)
}

// ListAccessibleReposRequest is what ListAccessibleRepos needs to ask a
// forge which repos a token can reach.
type ListAccessibleReposRequest struct {
	// InstanceURL is set for Forgejo, empty for GitHub.
	InstanceURL string
	Token       string
}

// ListRunsRequest is what ListRecentRuns needs to poll a tracked repo.
type ListRunsRequest struct {
	// InstanceURL is set for Forgejo, empty for GitHub.
	InstanceURL string
	Identifier  string
	Token       string
	// ETag is the repo's ReconcileETag from the last poll, sent as
	// If-None-Match. Empty on a repo's first poll.
	ETag string
}

// ListRunsResult is what a reconciliation poll found.
type ListRunsResult struct {
	// NotModified is true when the forge reported nothing changed since
	// ETag (a 304 on GitHub); Runs is empty and ignored in that case.
	NotModified bool
	// ETag is the new value to persist as the repo's ReconcileETag, for
	// If-None-Match on the next poll. Always empty on Forgejo.
	ETag string
	Runs []RunSnapshot
}

// RunSnapshot is a workflow run as reconciliation polling found it on the
// forge, with its jobs and their steps nested -- unlike Run/Job/Step, which
// are flat storage records keyed by ID once persisted.
type RunSnapshot struct {
	ForgeRunID   string
	PipelineName string
	Status       string
	Conclusion   string
	StartedAt    *time.Time
	CompletedAt  *time.Time
	ForgeURL     string
	Jobs         []JobSnapshot
}

// JobSnapshot is a workflow job nested under a RunSnapshot.
type JobSnapshot struct {
	ForgeJobID  string
	Name        string
	Status      string
	Conclusion  string
	QueuedAt    *time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	ForgeURL    string
	Steps       []StepSnapshot
}

// StepSnapshot is a job step nested under a JobSnapshot.
type StepSnapshot struct {
	Number      int
	Name        string
	Status      string
	Conclusion  string
	StartedAt   *time.Time
	CompletedAt *time.Time
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

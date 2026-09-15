// Package forgejo is the ingestion.ForgeClient adapter for Forgejo, backed
// by the Gitea SDK (Forgejo's REST API is Gitea-derived).
package forgejo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

// ErrInvalidIdentifier is returned when a repo identifier isn't "owner/name".
var ErrInvalidIdentifier = errors.New("identifier must be in owner/name form")

// Client is the Forgejo-backed ingestion.ForgeClient. A single Client
// serves any Forgejo instance URL, since each request carries its own.
type Client struct{}

// NewClient returns a Client.
func NewClient() (*Client, error) {
	return &Client{}, nil
}

// CreateWebhook implements ingestion.ForgeClient.
func (c *Client) CreateWebhook(ctx context.Context, req ingestion.CreateWebhookRequest) error {
	owner, name, err := splitIdentifier(req.Identifier)
	if err != nil {
		return err
	}

	api, err := gitea.NewClient(req.InstanceURL,
		gitea.SetToken(req.Token),
		gitea.SetContext(ctx),
		gitea.SetGiteaVersion(""), // skip the live version-check request
	)
	if err != nil {
		return fmt.Errorf("create forgejo client: %w", err)
	}

	_, _, err = api.CreateRepoHook(owner, name, gitea.CreateHookOption{
		Type: gitea.HookTypeGitea,
		Config: map[string]string{
			"url":          req.CallbackURL,
			"content_type": "json",
			"secret":       req.Secret,
		},
		// Forgejo Actions mirrors GitHub's workflow_run/workflow_job event
		// naming; unverified against a live instance (design.md's
		// "Gitea SDK's Forgejo Actions endpoint coverage is unverified" risk).
		Events: []string{"workflow_run", "workflow_job"},
		Active: true,
	})
	if err != nil {
		return fmt.Errorf("create forgejo webhook: %w", err)
	}

	return nil
}

func splitIdentifier(identifier string) (owner, name string, err error) {
	parts := strings.SplitN(identifier, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: %q", ErrInvalidIdentifier, identifier)
	}

	return parts[0], parts[1], nil
}

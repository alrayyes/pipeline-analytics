// Package github is the ingestion.ForgeClient adapter for GitHub, backed by
// google/go-github.
package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ghapi "github.com/google/go-github/v75/github"
)

// ErrInvalidIdentifier is returned when a repo identifier isn't "owner/name".
var ErrInvalidIdentifier = errors.New("identifier must be in owner/name form")

// Client is the GitHub-backed ingestion.ForgeClient.
type Client struct {
	baseURL *url.URL
}

// NewClient returns a Client. baseURL overrides the GitHub API endpoint
// (for tests against a fake server); pass "" to use the real GitHub API.
func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		return &Client{}, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	return &Client{baseURL: u}, nil
}

// CreateWebhook implements ingestion.ForgeClient.
func (c *Client) CreateWebhook(ctx context.Context, req ingestion.CreateWebhookRequest) error {
	owner, name, err := splitIdentifier(req.Identifier)
	if err != nil {
		return err
	}

	api := ghapi.NewClient(nil).WithAuthToken(req.Token)
	if c.baseURL != nil {
		api.BaseURL = c.baseURL
	}

	hook := &ghapi.Hook{
		Config: &ghapi.HookConfig{
			URL:         new(req.CallbackURL),
			ContentType: new("json"),
			Secret:      new(req.Secret),
		},
		Events: []string{"workflow_run", "workflow_job"},
		Active: new(true),
	}

	if _, _, err := api.Repositories.CreateHook(ctx, owner, name, hook); err != nil {
		return fmt.Errorf("create github webhook: %w", err)
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

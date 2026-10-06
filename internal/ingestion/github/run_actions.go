package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

var _ ingestion.RunActor = (*Client)(nil)

// RerunRun implements ingestion.RunActor. It re-runs only the failed jobs
// when asked to, the whole run otherwise
// (https://docs.github.com/en/rest/actions/workflow-runs).
func (c *Client) RerunRun(ctx context.Context, req ingestion.RunActionRequest) error {
	path := "rerun"
	if req.FailedOnly {
		path = "rerun-failed-jobs"
	}

	return c.postRunAction(ctx, req, path)
}

// CancelRun implements ingestion.RunActor.
func (c *Client) CancelRun(ctx context.Context, req ingestion.RunActionRequest) error {
	return c.postRunAction(ctx, req, "cancel")
}

func (c *Client) postRunAction(ctx context.Context, req ingestion.RunActionRequest, path string) error {
	owner, name, err := splitIdentifier(req.Identifier)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%srepos/%s/%s/actions/runs/%s/%s", c.baseURLString(), owner, name, req.ForgeRunID, path)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	httpReq.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %w", ingestion.ErrActionUnreachable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	c.recordRateLimit(req.Token, resp.Header)

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return ingestion.ErrActionForbidden
	case http.StatusNotFound:
		// GitHub answers 404 for a repo the token can't write to, too.
		return ingestion.ErrActionForbidden
	case http.StatusConflict:
		return ingestion.ErrNotActionable
	default:
		return fmt.Errorf("%w: status %d", ingestion.ErrActionUnreachable, resp.StatusCode)
	}
}

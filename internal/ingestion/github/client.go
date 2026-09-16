// Package github is the ingestion.ForgeClient adapter for GitHub, backed by
// google/go-github.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ghapi "github.com/google/go-github/v75/github"
)

// defaultBaseURL is GitHub's real API base, used when NewClient wasn't
// given a fake one to test against.
const defaultBaseURL = "https://api.github.com/"

// ErrInvalidIdentifier is returned when a repo identifier isn't "owner/name".
var ErrInvalidIdentifier = errors.New("identifier must be in owner/name form")

// ErrUnexpectedStatus is returned when a GitHub API response's status code
// is neither the expected success code nor 304 Not Modified.
var ErrUnexpectedStatus = errors.New("unexpected status")

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

// ListAccessibleRepos implements ingestion.ForgeClient. Returns the first
// 100 repos (owner, collaborator, and organization-member repos) the token
// can see, newest-pushed first -- enough for the registration UI's picker
// without adding pagination nobody's asked for yet.
func (c *Client) ListAccessibleRepos(ctx context.Context, req ingestion.ListAccessibleReposRequest) ([]string, error) {
	api := ghapi.NewClient(nil).WithAuthToken(req.Token)
	if c.baseURL != nil {
		api.BaseURL = c.baseURL
	}

	repos, _, err := api.Repositories.ListByAuthenticatedUser(ctx, &ghapi.RepositoryListByAuthenticatedUserOptions{
		Sort:    "pushed",
		PerPage: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("list accessible github repos: %w", err)
	}

	identifiers := make([]string, 0, len(repos))
	for _, r := range repos {
		identifiers = append(identifiers, r.GetFullName())
	}

	return identifiers, nil
}

// ListRecentRuns implements ingestion.ForgeClient. The run-list request is
// conditional (If-None-Match/ETag); a 304 short-circuits before any jobs
// are fetched, since nothing about the run list changed. go-github's typed
// client isn't used here -- CheckResponse treats a 304 as an API error, so
// a plain net/http request keeps the conditional-GET contract explicit and
// testable against a fake server.
func (c *Client) ListRecentRuns(ctx context.Context, req ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
	owner, name, err := splitIdentifier(req.Identifier)
	if err != nil {
		return ingestion.ListRunsResult{}, err
	}

	var runsPayload struct {
		WorkflowRuns []struct {
			ID           int64  `json:"id"`
			Name         string `json:"name"`
			Status       string `json:"status"`
			Conclusion   string `json:"conclusion"`
			RunStartedAt string `json:"run_started_at"`
			UpdatedAt    string `json:"updated_at"`
			HTMLURL      string `json:"html_url"`
		} `json:"workflow_runs"`
	}

	runsURL := fmt.Sprintf("%srepos/%s/%s/actions/runs?per_page=20", c.baseURLString(), owner, name)

	etag, notModified, err := c.conditionalGet(ctx, runsURL, req.Token, req.ETag, &runsPayload)
	if err != nil {
		return ingestion.ListRunsResult{}, fmt.Errorf("list workflow runs: %w", err)
	}

	if notModified {
		return ingestion.ListRunsResult{NotModified: true}, nil
	}

	runs := make([]ingestion.RunSnapshot, 0, len(runsPayload.WorkflowRuns))

	for _, wr := range runsPayload.WorkflowRuns {
		jobs, err := c.listWorkflowJobs(ctx, owner, name, wr.ID, req.Token)
		if err != nil {
			return ingestion.ListRunsResult{}, err
		}

		var completedAt *time.Time
		if wr.Status == "completed" {
			completedAt = parseGitHubTime(wr.UpdatedAt)
		}

		runs = append(runs, ingestion.RunSnapshot{
			ForgeRunID:   strconv.FormatInt(wr.ID, 10),
			PipelineName: wr.Name,
			Status:       wr.Status,
			Conclusion:   wr.Conclusion,
			StartedAt:    parseGitHubTime(wr.RunStartedAt),
			CompletedAt:  completedAt,
			ForgeURL:     wr.HTMLURL,
			Jobs:         jobs,
		})
	}

	return ingestion.ListRunsResult{ETag: etag, Runs: runs}, nil
}

func (c *Client) listWorkflowJobs(ctx context.Context, owner, name string, runID int64, token string) ([]ingestion.JobSnapshot, error) {
	var payload struct {
		Jobs []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			Status      string `json:"status"`
			Conclusion  string `json:"conclusion"`
			CreatedAt   string `json:"created_at"`
			StartedAt   string `json:"started_at"`
			CompletedAt string `json:"completed_at"`
			HTMLURL     string `json:"html_url"`
			Steps       []struct {
				Name        string `json:"name"`
				Status      string `json:"status"`
				Conclusion  string `json:"conclusion"`
				Number      int    `json:"number"`
				StartedAt   string `json:"started_at"`
				CompletedAt string `json:"completed_at"`
			} `json:"steps"`
		} `json:"jobs"`
	}

	jobsURL := fmt.Sprintf("%srepos/%s/%s/actions/runs/%d/jobs", c.baseURLString(), owner, name, runID)

	if _, _, err := c.conditionalGet(ctx, jobsURL, token, "", &payload); err != nil {
		return nil, fmt.Errorf("list workflow jobs for run %d: %w", runID, err)
	}

	jobs := make([]ingestion.JobSnapshot, 0, len(payload.Jobs))

	for _, j := range payload.Jobs {
		steps := make([]ingestion.StepSnapshot, 0, len(j.Steps))
		for _, s := range j.Steps {
			steps = append(steps, ingestion.StepSnapshot{
				Number:      s.Number,
				Name:        s.Name,
				Status:      s.Status,
				Conclusion:  s.Conclusion,
				StartedAt:   parseGitHubTime(s.StartedAt),
				CompletedAt: parseGitHubTime(s.CompletedAt),
			})
		}

		jobs = append(jobs, ingestion.JobSnapshot{
			ForgeJobID:  strconv.FormatInt(j.ID, 10),
			Name:        j.Name,
			Status:      j.Status,
			Conclusion:  j.Conclusion,
			QueuedAt:    parseGitHubTime(j.CreatedAt),
			StartedAt:   parseGitHubTime(j.StartedAt),
			CompletedAt: parseGitHubTime(j.CompletedAt),
			ForgeURL:    j.HTMLURL,
			Steps:       steps,
		})
	}

	return jobs, nil
}

// conditionalGet issues an authenticated GET, sending ifNoneMatch as
// If-None-Match when set, and JSON-decodes a 200 response into out. A 304
// reports notModified with no decode attempted.
func (c *Client) conditionalGet(ctx context.Context, url, token, ifNoneMatch string, out any) (etag string, notModified bool, err error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", false, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Accept", "application/vnd.github+json")

	if ifNoneMatch != "" {
		httpReq.Header.Set("If-None-Match", ifNoneMatch)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", false, fmt.Errorf("request %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotModified {
		return "", true, nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("%s: %w: %d", url, ErrUnexpectedStatus, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return "", false, fmt.Errorf("decode response from %s: %w", url, err)
	}

	return resp.Header.Get("ETag"), false, nil
}

func (c *Client) baseURLString() string {
	if c.baseURL != nil {
		return c.baseURL.String()
	}

	return defaultBaseURL
}

func parseGitHubTime(s string) *time.Time {
	if s == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}

	return &t
}

func splitIdentifier(identifier string) (owner, name string, err error) {
	parts := strings.SplitN(identifier, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: %q", ErrInvalidIdentifier, identifier)
	}

	return parts[0], parts[1], nil
}

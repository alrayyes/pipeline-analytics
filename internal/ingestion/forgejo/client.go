// Package forgejo is the ingestion.ForgeClient adapter for Forgejo, backed
// by the Gitea SDK (Forgejo's REST API is Gitea-derived).
package forgejo

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

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

// ListAccessibleRepos implements ingestion.ForgeClient. Returns every repo
// the token can see, excluding archived and forked repos -- neither is
// something the registration UI's picker should ever offer to track.
func (c *Client) ListAccessibleRepos(ctx context.Context, req ingestion.ListAccessibleReposRequest) ([]string, error) {
	api, err := gitea.NewClient(req.InstanceURL,
		gitea.SetToken(req.Token),
		gitea.SetContext(ctx),
		gitea.SetGiteaVersion(""), // skip the live version-check request
	)
	if err != nil {
		return nil, fmt.Errorf("create forgejo client: %w", err)
	}

	opts := gitea.ListReposOptions{}
	var identifiers []string

	for {
		repos, resp, err := api.ListMyRepos(opts)
		if err != nil {
			return nil, fmt.Errorf("list accessible forgejo repos: %w", err)
		}

		for _, r := range repos {
			if r.Archived || r.Fork {
				continue
			}

			identifiers = append(identifiers, r.FullName)
		}

		if resp.NextPage == 0 {
			return identifiers, nil
		}

		opts.Page = resp.NextPage
	}
}

// ListRecentRuns implements ingestion.ForgeClient. Forgejo has no default
// rate limit and no documented conditional-request support for this
// endpoint (design.md's "Ingestion" decision), so every poll fetches the
// full recent run list -- NotModified and ETag are always unset.
func (c *Client) ListRecentRuns(ctx context.Context, req ingestion.ListRunsRequest) (ingestion.ListRunsResult, error) {
	owner, name, err := splitIdentifier(req.Identifier)
	if err != nil {
		return ingestion.ListRunsResult{}, err
	}

	api, err := gitea.NewClient(req.InstanceURL,
		gitea.SetToken(req.Token),
		gitea.SetContext(ctx),
		gitea.SetGiteaVersion(""), // skip the live version-check request
	)
	if err != nil {
		return ingestion.ListRunsResult{}, fmt.Errorf("create forgejo client: %w", err)
	}

	runsResp, _, err := api.ListRepoActionRuns(owner, name, gitea.ListRepoActionRunsOptions{})
	if err != nil {
		return ingestion.ListRunsResult{}, fmt.Errorf("list forgejo workflow runs: %w", err)
	}

	runs := make([]ingestion.RunSnapshot, 0, len(runsResp.WorkflowRuns))

	for _, run := range runsResp.WorkflowRuns {
		jobs, err := c.listWorkflowJobs(api, owner, name, run.ID)
		if err != nil {
			return ingestion.ListRunsResult{}, err
		}

		runs = append(runs, ingestion.RunSnapshot{
			ForgeRunID:   strconv.FormatInt(run.ID, 10),
			PipelineName: workflowNameFromPath(run.Path),
			Status:       run.Status,
			Conclusion:   run.Conclusion,
			StartedAt:    nonZeroTime(run.StartedAt),
			CompletedAt:  nonZeroTime(run.CompletedAt),
			ForgeURL:     run.HTMLURL,
			Jobs:         jobs,
		})
	}

	return ingestion.ListRunsResult{Runs: runs}, nil
}

func (c *Client) listWorkflowJobs(api *gitea.Client, owner, name string, runID int64) ([]ingestion.JobSnapshot, error) {
	jobsResp, _, err := api.ListRepoActionRunJobs(owner, name, runID, gitea.ListRepoActionJobsOptions{})
	if err != nil {
		return nil, fmt.Errorf("list forgejo workflow jobs for run %d: %w", runID, err)
	}

	jobs := make([]ingestion.JobSnapshot, 0, len(jobsResp.Jobs))

	for _, job := range jobsResp.Jobs {
		steps := make([]ingestion.StepSnapshot, 0, len(job.Steps))
		for _, step := range job.Steps {
			steps = append(steps, ingestion.StepSnapshot{
				Number:      int(step.Number),
				Name:        step.Name,
				Status:      step.Status,
				Conclusion:  step.Conclusion,
				StartedAt:   nonZeroTime(step.StartedAt),
				CompletedAt: nonZeroTime(step.CompletedAt),
			})
		}

		jobs = append(jobs, ingestion.JobSnapshot{
			ForgeJobID:  strconv.FormatInt(job.ID, 10),
			Name:        job.Name,
			Status:      job.Status,
			Conclusion:  job.Conclusion,
			QueuedAt:    nonZeroTime(job.CreatedAt),
			StartedAt:   nonZeroTime(job.StartedAt),
			CompletedAt: nonZeroTime(job.CompletedAt),
			ForgeURL:    job.HTMLURL,
			Steps:       steps,
		})
	}

	return jobs, nil
}

// workflowNameFromPath derives a display name from a workflow file's
// repo-relative path (e.g. ".gitea/workflows/ci.yml" -> "ci") --
// ActionWorkflowRun carries no separate workflow name field.
func workflowNameFromPath(workflowPath string) string {
	base := path.Base(workflowPath)

	return strings.TrimSuffix(strings.TrimSuffix(base, ".yaml"), ".yml")
}

func nonZeroTime(t time.Time) *time.Time {
	if t.IsZero() {
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

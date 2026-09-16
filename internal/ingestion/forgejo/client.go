// Package forgejo is the ingestion.ForgeClient adapter for Forgejo, backed
// by the Gitea SDK (Forgejo's REST API is Gitea-derived).
package forgejo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
		// Forgejo has no webhook event for Actions run/job status changes
		// as of 11.0.16+gitea-1.22.0 -- confirmed empirically, not just
		// undocumented: walked the live "new hook" form's checkbox names
		// (the full vocabulary is create/delete/fork/issues/issue_assign/
		// issue_comment/issue_label/issue_milestone/package/pull_request/
		// pull_request_assign/pull_request_comment/pull_request_label/
		// pull_request_milestone/pull_request_review/
		// pull_request_review_request/pull_request_sync/push/release/
		// repository/wiki, nothing Actions-related), and separately
		// confirmed "workflow_run"/"workflow_job" both silently drop from
		// a created hook's events (github.com/alrayyes/pipeline-analytics
		// #120, citing alrayyes/forge-dashboard#177's earlier finding on
		// a different instance). "push" is the closest available signal
		// that a workflow run may be starting; ingestion.ProcessForgejoEvent
		// treats delivery as a cue to reconcile the repo immediately
		// rather than parsed run data, since a push payload carries none.
		Events: []string{"push"},
		Active: true,
	})
	if err != nil {
		return fmt.Errorf("create forgejo webhook: %w", err)
	}

	return nil
}

// ListAccessibleRepos implements ingestion.ForgeClient. Returns every repo
// the token can see, excluding archived, forked, and mirror repos -- none
// of the three is something the registration UI's picker should ever offer
// to track (a mirror commonly has no Actions runs of its own to reconcile,
// e.g. a Forgejo repo that just mirrors a GitHub one).
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
			if r.Archived || r.Fork || r.Mirror {
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

	runsResp, resp, err := api.ListRepoActionRuns(owner, name, gitea.ListRepoActionRunsOptions{})
	if err != nil {
		// Forgejo 404s this whole route -- "The target couldn't be
		// found." -- for a repo with Actions disabled or that has never
		// had a workflow run, rather than returning a 200 with an empty
		// list (confirmed against real deployments,
		// github.com/alrayyes/pipeline-analytics#121). That's not a
		// reconciliation failure worth an error log every poll; it's a
		// legitimate "nothing to reconcile" for a repo someone tracked
		// that simply has no pipelines (yet).
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return ingestion.ListRunsResult{}, nil
		}

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

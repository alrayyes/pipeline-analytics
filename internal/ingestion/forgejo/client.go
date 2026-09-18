// Package forgejo is the ingestion.ForgeClient adapter for Forgejo, backed
// by the Gitea SDK (Forgejo's REST API is Gitea-derived).
package forgejo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// ErrUnexpectedStatus is returned when a Forgejo API response's status code
// is neither success nor a recognized "nothing to report" case.
var ErrUnexpectedStatus = errors.New("unexpected status")

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

// GetRepo implements ingestion.ForgeClient. A single-repo lookup, not
// ListAccessibleRepos filtered client-side -- the authoritative check
// registration needs, independent of whatever a prior Discover call saw
// (and of whether the bulk listing endpoint populates these flags as
// reliably as a single-repo fetch does).
func (c *Client) GetRepo(ctx context.Context, req ingestion.GetRepoRequest) (ingestion.RepoMetadata, error) {
	owner, name, err := splitIdentifier(req.Identifier)
	if err != nil {
		return ingestion.RepoMetadata{}, err
	}

	api, err := gitea.NewClient(req.InstanceURL,
		gitea.SetToken(req.Token),
		gitea.SetContext(ctx),
		gitea.SetGiteaVersion(""), // skip the live version-check request
	)
	if err != nil {
		return ingestion.RepoMetadata{}, fmt.Errorf("create forgejo client: %w", err)
	}

	r, _, err := api.GetRepo(owner, name)
	if err != nil {
		return ingestion.RepoMetadata{}, fmt.Errorf("get forgejo repo: %w", err)
	}

	return ingestion.RepoMetadata{
		Archived: r.Archived,
		Fork:     r.Fork,
		Mirror:   r.Mirror,
	}, nil
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
		jobs, err := c.listWorkflowJobs(ctx, req.InstanceURL, req.Token, owner, name, run.ID)
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

// listWorkflowJobs fetches a run's jobs with a raw HTTP request rather than
// the Gitea SDK's typed ListRepoActionRunJobs -- that method decodes into
// ActionWorkflowJobsResponse ({"jobs": [...]}), a shape confirmed against a
// real deployment to not match what Forgejo actually returns there (a bare
// array, github.com/alrayyes/pipeline-analytics#123). parseWorkflowJobs
// accepts either shape, so this works whichever one a given Forgejo version
// sends. Also confirmed empirically (Forgejo 11 through 14, every currently
// tagged release) that this endpoint isn't registered at all on some
// versions -- a 404 there is treated as "no job detail available" rather
// than failing the whole run, same spirit as the runs-endpoint 404 handling
// above (#121).
func (c *Client) listWorkflowJobs(ctx context.Context, instanceURL, token, owner, name string, runID int64) ([]ingestion.JobSnapshot, error) {
	jobsResp, err := fetchWorkflowJobs(ctx, instanceURL, token, owner, name, runID)
	if err != nil {
		return nil, err
	}

	jobs := make([]ingestion.JobSnapshot, 0, len(jobsResp))
	for _, job := range jobsResp {
		jobs = append(jobs, convertWorkflowJob(job))
	}

	return jobs, nil
}

func fetchWorkflowJobs(ctx context.Context, instanceURL, token, owner, name string, runID int64) ([]*gitea.ActionWorkflowJob, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/actions/runs/%d/jobs", strings.TrimRight(instanceURL, "/"), owner, name, runID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build forgejo jobs request for run %d: %w", runID, err)
	}

	httpReq.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("list forgejo workflow jobs for run %d: %w", runID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read forgejo workflow jobs response for run %d: %w", runID, err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("list forgejo workflow jobs for run %d: %w: %d: %s", runID, ErrUnexpectedStatus, resp.StatusCode, body)
	}

	jobs, err := parseWorkflowJobs(body)
	if err != nil {
		return nil, fmt.Errorf("parse forgejo workflow jobs response for run %d: %w", runID, err)
	}

	return jobs, nil
}

// parseWorkflowJobs accepts either shape a Forgejo instance's jobs-listing
// endpoint might return: a bare array (confirmed against a real deployment,
// #123) or the {"jobs": [...]} wrapper the Gitea SDK's own
// ActionWorkflowJobsResponse type expects and upstream Gitea documents.
func parseWorkflowJobs(body []byte) ([]*gitea.ActionWorkflowJob, error) {
	var bare []*gitea.ActionWorkflowJob
	if err := json.Unmarshal(body, &bare); err == nil {
		return bare, nil
	}

	var wrapped gitea.ActionWorkflowJobsResponse
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, fmt.Errorf("unmarshal forgejo workflow jobs: %w", err)
	}

	return wrapped.Jobs, nil
}

func convertWorkflowJob(job *gitea.ActionWorkflowJob) ingestion.JobSnapshot {
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

	return ingestion.JobSnapshot{
		ForgeJobID:  strconv.FormatInt(job.ID, 10),
		Name:        job.Name,
		Status:      job.Status,
		Conclusion:  job.Conclusion,
		QueuedAt:    nonZeroTime(job.CreatedAt),
		StartedAt:   nonZeroTime(job.StartedAt),
		CompletedAt: nonZeroTime(job.CompletedAt),
		ForgeURL:    job.HTMLURL,
		Steps:       steps,
	}
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

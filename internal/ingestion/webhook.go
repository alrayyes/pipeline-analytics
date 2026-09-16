package ingestion

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// VerifySignature reports whether signature is a valid HMAC-SHA256 of body
// under secret. Accepts either GitHub's "sha256=<hex>" form or a bare hex
// digest (Forgejo's format, unverified against a live instance -- see
// design.md's "Gitea SDK's Forgejo Actions endpoint coverage is
// unverified" risk).
func VerifySignature(secret string, body []byte, signature string) bool {
	digest := strings.TrimPrefix(signature, "sha256=")

	given, err := hex.DecodeString(digest)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	want := mac.Sum(nil)

	return subtle.ConstantTimeCompare(given, want) == 1
}

// ProcessGitHubEvent updates run/job/step storage for a single GitHub
// Actions webhook delivery already verified and resolved to repo. Event
// types other than workflow_run and workflow_job are ignored, not an error
// -- GitHub's own webhook config on a repo may carry more event types than
// this project subscribes to on purpose.
func ProcessGitHubEvent(ctx context.Context, store RunStore, repo Repo, eventType string, payload []byte) error {
	switch eventType {
	case "workflow_run":
		return processGitHubWorkflowRun(ctx, store, repo, payload)
	case "workflow_job":
		return processGitHubWorkflowJob(ctx, store, repo, payload)
	default:
		return nil
	}
}

// ProcessForgejoEvent handles a single Forgejo webhook delivery already
// verified and resolved to repo. Forgejo has no webhook event for Actions
// run/job status changes as of 11.0.16+gitea-1.22.0 -- confirmed
// empirically against a live container (github.com/alrayyes/pipeline-
// analytics#120), not just undocumented, so unlike GitHub there's no
// run/job payload to parse here. The webhook forgejoclient.CreateWebhook
// registers subscribes to "push" only, the closest available signal that a
// workflow run may be starting; delivery of one triggers an immediate
// reconciliation poll for the repo instead.
func ProcessForgejoEvent(ctx context.Context, reconciler RepoReconciler, repo Repo, eventType string) error {
	if eventType != "push" {
		return nil
	}

	if err := reconciler.ReconcileRepo(ctx, repo); err != nil {
		return fmt.Errorf("reconcile repo: %w", err)
	}

	return nil
}

type githubWorkflowRunEvent struct {
	WorkflowRun struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Status       string `json:"status"`
		Conclusion   string `json:"conclusion"`
		RunStartedAt string `json:"run_started_at"`
		UpdatedAt    string `json:"updated_at"`
		HTMLURL      string `json:"html_url"`
	} `json:"workflow_run"`
}

func processGitHubWorkflowRun(ctx context.Context, store RunStore, repo Repo, payload []byte) error {
	var event githubWorkflowRunEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal workflow_run event: %w", err)
	}

	run := Run{
		RepoID:       repo.ID,
		ForgeRunID:   strconv.FormatInt(event.WorkflowRun.ID, 10),
		PipelineName: event.WorkflowRun.Name,
		Status:       event.WorkflowRun.Status,
		Conclusion:   event.WorkflowRun.Conclusion,
		StartedAt:    parseTime(event.WorkflowRun.RunStartedAt),
		ForgeURL:     event.WorkflowRun.HTMLURL,
	}

	if run.Status == "completed" {
		run.CompletedAt = parseTime(event.WorkflowRun.UpdatedAt)
	}

	if _, err := store.UpsertRun(ctx, run); err != nil {
		return fmt.Errorf("upsert run: %w", err)
	}

	return nil
}

type githubWorkflowJobEvent struct {
	WorkflowJob struct {
		ID           int64           `json:"id"`
		RunID        int64           `json:"run_id"`
		WorkflowName string          `json:"workflow_name"`
		Name         string          `json:"name"`
		Status       string          `json:"status"`
		Conclusion   string          `json:"conclusion"`
		CreatedAt    string          `json:"created_at"`
		StartedAt    string          `json:"started_at"`
		CompletedAt  string          `json:"completed_at"`
		HTMLURL      string          `json:"html_url"`
		Steps        []githubJobStep `json:"steps"`
	} `json:"workflow_job"`
}

type githubJobStep struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	Number      int    `json:"number"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

func processGitHubWorkflowJob(ctx context.Context, store RunStore, repo Repo, payload []byte) error {
	var event githubWorkflowJobEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal workflow_job event: %w", err)
	}

	// The parent run may not have arrived yet (webhook delivery order isn't
	// guaranteed) -- upsert a stub with what the job payload knows; a later
	// workflow_run event, or reconciliation, fills in the rest.
	run, err := store.UpsertRun(ctx, Run{
		RepoID:       repo.ID,
		ForgeRunID:   strconv.FormatInt(event.WorkflowJob.RunID, 10),
		PipelineName: event.WorkflowJob.WorkflowName,
		Status:       "in_progress",
	})
	if err != nil {
		return fmt.Errorf("upsert parent run: %w", err)
	}

	job := Job{
		RunID:       run.ID,
		ForgeJobID:  strconv.FormatInt(event.WorkflowJob.ID, 10),
		Name:        event.WorkflowJob.Name,
		Status:      event.WorkflowJob.Status,
		Conclusion:  event.WorkflowJob.Conclusion,
		QueuedAt:    parseTime(event.WorkflowJob.CreatedAt),
		StartedAt:   parseTime(event.WorkflowJob.StartedAt),
		CompletedAt: parseTime(event.WorkflowJob.CompletedAt),
		ForgeURL:    event.WorkflowJob.HTMLURL,
	}

	storedJob, err := store.UpsertJob(ctx, job)
	if err != nil {
		return fmt.Errorf("upsert job: %w", err)
	}

	steps := make([]Step, 0, len(event.WorkflowJob.Steps))
	for _, s := range event.WorkflowJob.Steps {
		steps = append(steps, Step{
			Number:      s.Number,
			Name:        s.Name,
			Status:      s.Status,
			Conclusion:  s.Conclusion,
			StartedAt:   parseTime(s.StartedAt),
			CompletedAt: parseTime(s.CompletedAt),
		})
	}

	if err := store.ReplaceSteps(ctx, storedJob.ID, steps); err != nil {
		return fmt.Errorf("replace steps: %w", err)
	}

	return nil
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}

	return &t
}

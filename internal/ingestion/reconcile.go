package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Reconciler polls each tracked repo's recent workflow runs on a schedule,
// per forge-ingestion/spec.md's "Reconciliation polling" requirement: it
// backfills a newly tracked repo's history and catches any webhook delivery
// that was missed, at no rate-limit cost when nothing has changed.
type Reconciler struct {
	store    Store
	runStore RunStore
	clients  map[Forge]ForgeClient
}

// NewReconciler returns a Reconciler.
func NewReconciler(store Store, runStore RunStore, clients map[Forge]ForgeClient) *Reconciler {
	return &Reconciler{store: store, runStore: runStore, clients: clients}
}

// Run polls every tracked repo immediately, then again every interval,
// until ctx is done.
func (r *Reconciler) Run(ctx context.Context, interval time.Duration) {
	r.ReconcileAll(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.ReconcileAll(ctx)
		}
	}
}

// ReconcileAll polls every tracked repo once. A single repo's failure is
// logged, not returned, so it doesn't block reconciling the others.
func (r *Reconciler) ReconcileAll(ctx context.Context) {
	repos, err := r.store.ListRepos(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "list tracked repos for reconciliation", "error", err)

		return
	}

	for _, repo := range repos {
		if err := r.ReconcileRepo(ctx, repo); err != nil {
			slog.ErrorContext(ctx, "reconcile repo", "repo", repo.Identifier, "error", err)
		}
	}
}

// ReconcileRepo polls one tracked repo and updates run/job/step storage
// with whatever the forge reports changed. A repo whose forge has no
// registered client is skipped rather than treated as an error, so a
// partially configured deployment doesn't fail every poll.
func (r *Reconciler) ReconcileRepo(ctx context.Context, repo Repo) error {
	client, ok := r.clients[repo.Forge]
	if !ok {
		return nil
	}

	token, err := r.store.RepoToken(ctx, repo.ID)
	if err != nil {
		return fmt.Errorf("load repo token: %w", err)
	}

	result, err := client.ListRecentRuns(ctx, ListRunsRequest{
		InstanceURL: repo.ForgejoInstanceURL,
		Identifier:  repo.Identifier,
		Token:       token,
		ETag:        repo.ReconcileETag,
	})
	if err != nil {
		return fmt.Errorf("list recent runs: %w", err)
	}

	if result.NotModified {
		return nil
	}

	for _, snapshot := range result.Runs {
		if err := r.storeRunSnapshot(ctx, repo, snapshot); err != nil {
			return err
		}
	}

	if result.ETag == "" {
		return nil
	}

	if err := r.store.SetReconcileETag(ctx, repo.ID, result.ETag); err != nil {
		return fmt.Errorf("persist reconciliation etag: %w", err)
	}

	return nil
}

func (r *Reconciler) storeRunSnapshot(ctx context.Context, repo Repo, snapshot RunSnapshot) error {
	run, err := r.runStore.UpsertRun(ctx, Run{
		RepoID:       repo.ID,
		ForgeRunID:   snapshot.ForgeRunID,
		PipelineName: snapshot.PipelineName,
		Status:       snapshot.Status,
		Conclusion:   snapshot.Conclusion,
		StartedAt:    snapshot.StartedAt,
		CompletedAt:  snapshot.CompletedAt,
		ForgeURL:     snapshot.ForgeURL,
	})
	if err != nil {
		return fmt.Errorf("upsert run: %w", err)
	}

	for _, job := range snapshot.Jobs {
		if err := r.storeJobSnapshot(ctx, run.ID, job); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) storeJobSnapshot(ctx context.Context, runID string, snapshot JobSnapshot) error {
	job, err := r.runStore.UpsertJob(ctx, Job{
		RunID:       runID,
		ForgeJobID:  snapshot.ForgeJobID,
		Name:        snapshot.Name,
		Status:      snapshot.Status,
		Conclusion:  snapshot.Conclusion,
		QueuedAt:    snapshot.QueuedAt,
		StartedAt:   snapshot.StartedAt,
		CompletedAt: snapshot.CompletedAt,
		ForgeURL:    snapshot.ForgeURL,
	})
	if err != nil {
		return fmt.Errorf("upsert job: %w", err)
	}

	steps := make([]Step, 0, len(snapshot.Steps))
	for _, s := range snapshot.Steps {
		steps = append(steps, Step{
			Number:      s.Number,
			Name:        s.Name,
			Status:      s.Status,
			Conclusion:  s.Conclusion,
			StartedAt:   s.StartedAt,
			CompletedAt: s.CompletedAt,
		})
	}

	if err := r.runStore.ReplaceSteps(ctx, job.ID, steps); err != nil {
		return fmt.Errorf("replace steps: %w", err)
	}

	return nil
}

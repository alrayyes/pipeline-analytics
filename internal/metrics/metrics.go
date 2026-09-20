// Package metrics derives health scores and pain-point signals from
// ingested pipeline run data. See
// openspec/changes/add-pipeline-dashboard/specs/pipeline-metrics/spec.md.
package metrics

import (
	"cmp"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HealthStatus is a pipeline's computed health, per pipeline-metrics/
// spec.md's "Pipeline health status" requirement.
type HealthStatus string

// The health statuses a pipeline can be reported as.
const (
	HealthHealthy   HealthStatus = "healthy"
	HealthUnhealthy HealthStatus = "unhealthy"
)

// Signal identifies which input triggered an unhealthy status.
type Signal string

// The signals that can trigger an unhealthy status.
const (
	SignalFailureRate        Signal = "failure_rate"
	SignalDurationRegression Signal = "duration_regression"
	SignalFlakyStep          Signal = "flaky_step"
)

// Threshold constants for health-status computation. design.md flags the
// exact formula/weighting as an open question, tunable after
// implementation without changing the underlying per-signal specs; these
// are a defensible starting point, not a final answer.
const (
	unhealthyFailureRateThreshold = 0.2
	durationRegressionFactor      = 1.5
	trendBucketSize               = 5
)

// PipelineRef identifies a pipeline: a distinct workflow name within a
// tracked repo. A pipeline has no storage row of its own -- it's derived
// from the distinct (repo, name) pairs already present in run history.
type PipelineRef struct {
	RepoID string
	Name   string
}

// PipelineID is a PipelineRef, opaque-encoded for use in a URL path and API
// response.
type PipelineID string

const pipelineIDSeparator = "\x00"

// ID derives r's opaque identifier.
func (r PipelineRef) ID() PipelineID {
	return PipelineID(base64.RawURLEncoding.EncodeToString([]byte(r.RepoID + pipelineIDSeparator + r.Name)))
}

// ErrInvalidPipelineID is returned when a PipelineID doesn't decode to a
// valid PipelineRef.
var ErrInvalidPipelineID = errors.New("invalid pipeline id")

// ParsePipelineID recovers the PipelineRef a PipelineID was derived from.
func ParsePipelineID(id PipelineID) (PipelineRef, error) {
	raw, err := base64.RawURLEncoding.DecodeString(string(id))
	if err != nil {
		return PipelineRef{}, fmt.Errorf("%w: %s", ErrInvalidPipelineID, id)
	}

	parts := strings.SplitN(string(raw), pipelineIDSeparator, 2)
	if len(parts) != 2 {
		return PipelineRef{}, fmt.Errorf("%w: %s", ErrInvalidPipelineID, id)
	}

	return PipelineRef{RepoID: parts[0], Name: parts[1]}, nil
}

// Window bounds how much run history a computation considers, as a
// trailing run count. The API's "window" query parameter also allows a
// duration, per its description; only a run count is implemented in v1 --
// a documented scope trim, not an oversight, since every requirement in
// pipeline-metrics/spec.md phrases a window in terms of "recent runs".
type Window struct {
	RunCount int
}

// DefaultWindowRunCount is used when no window is requested.
const DefaultWindowRunCount = 20

// ParseWindow interprets raw (the API's window query parameter). An empty
// or unrecognized value falls back to DefaultWindowRunCount.
func ParseWindow(raw string) Window {
	if n, err := strconv.Atoi(raw); err == nil && n > 0 {
		return Window{RunCount: n}
	}

	return Window{RunCount: DefaultWindowRunCount}
}

// doubled is window's own size plus the equally sized window immediately
// before it, for regression detection (current vs. prior).
func (w Window) doubled() Window {
	n := w.RunCount
	if n <= 0 {
		n = DefaultWindowRunCount
	}

	return Window{RunCount: n * 2}
}

// Pipeline is one tracked pipeline's identity and current health.
type Pipeline struct {
	ID               PipelineID
	RepoID           string
	Name             string
	HealthStatus     HealthStatus
	TriggeredSignals []Signal
	// LastRunAt is the most recent run's start time, nil if it has none.
	LastRunAt *time.Time
}

// Trend is a time series: parallel Timestamps/P50/P90 (a duration trend) or
// Timestamps/Rate (a failure-rate trend).
type Trend struct {
	Timestamps []time.Time
	P50        []float64
	P90        []float64
	Rate       []float64
}

// PipelineDetail is a pipeline plus its duration and failure-rate trends.
type PipelineDetail struct {
	Pipeline
	DurationTrend    Trend
	FailureRateTrend Trend
}

// Step is one named step's aggregated metrics within a pipeline.
type Step struct {
	ID                          string
	Name                        string
	DurationContributionSeconds float64
	QueueSeconds                float64
	ExecSeconds                 float64
	FailureRate                 float64
	FailureCount                int
	Flaky                       bool
	ForgeURL                    string
}

// FlakyRun is one run in which a specific step failed -- the step-scoped
// drill-down from a flaky step, so a click lands on a run it actually
// failed on instead of an arbitrarily-picked occurrence.
type FlakyRun struct {
	RunID     string
	StartedAt *time.Time
	ForgeURL  string
}

// RunStep is one step's status within a single run, for the drill-down a
// FlakyRun leads to.
type RunStep struct {
	Name       string
	Status     string
	Conclusion string
	ForgeURL   string
}

// RunDetail is one run's own steps, in recorded order.
type RunDetail struct {
	RunID     string
	StartedAt *time.Time
	Steps     []RunStep
}

// UsageEntry is runner-minutes consumed by one workflow.
type UsageEntry struct {
	Workflow      string
	RunnerMinutes float64
}

// RunRecord is one pipeline run, as metrics computation needs it.
type RunRecord struct {
	ID          string
	Status      string
	Conclusion  string
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// StepOccurrence is one recorded execution of a named step, within one job
// of one run. JobQueuedAt/JobStartedAt bound the parent job's own queue
// wait -- shared by every step within it, since a step can't be
// individually queued.
type StepOccurrence struct {
	Name         string
	Status       string
	Conclusion   string
	StartedAt    *time.Time
	CompletedAt  *time.Time
	JobQueuedAt  *time.Time
	JobStartedAt *time.Time
	JobForgeURL  string
	RunID        string
	RunStartedAt *time.Time
}

// UsageRecord is one job's execution duration, attributed to its pipeline
// (workflow) name.
type UsageRecord struct {
	PipelineName string
	ExecSeconds  float64
}

// PipelineListFilter narrows ListPipelines to one repo and/or forge, and/or
// one page, ordered deterministically by (RepoID, Name). The zero value
// matches every tracked pipeline, unfiltered and unpaginated -- what
// ListUnhealthySteps needs, since it operates over every pipeline rather
// than one page a user is looking at. Mirrors ingestion.RepoListFilter.
type PipelineListFilter struct {
	// RepoID restricts the list to one tracked repo. Empty matches every repo.
	RepoID string
	// Forge restricts the list to one forge, via a join on the owning repo
	// (a pipeline has no forge column of its own). Empty matches every forge.
	Forge string
	// Limit caps how many pipelines are returned. Zero means unlimited.
	Limit int
	// Offset skips this many matching pipelines before the page starts.
	Offset int
}

// Store is the port metrics computations read run/job/step history
// through. Every method that returns runs orders them most-recent-first.
type Store interface {
	// ListPipelines returns the distinct pipelines matching filter, plus
	// whether more pipelines beyond this page also match -- always false
	// when filter.Limit is 0.
	ListPipelines(ctx context.Context, filter PipelineListFilter) (pipelines []PipelineRef, hasMore bool, err error)
	// PipelineRuns returns a pipeline's runs, most-recent-first, limited to
	// window.RunCount.
	PipelineRuns(ctx context.Context, ref PipelineRef, window Window) ([]RunRecord, error)
	// PipelineSteps returns every step occurrence recorded across a
	// pipeline's jobs within window, one entry per occurrence (not
	// pre-aggregated).
	PipelineSteps(ctx context.Context, ref PipelineRef, window Window) ([]StepOccurrence, error)
	// RepoUsage returns every job's execution duration within window for a
	// tracked repo, one entry per job.
	RepoUsage(ctx context.Context, repoID string, window Window) ([]UsageRecord, error)
	// RunSteps returns every step occurrence recorded within one run's
	// jobs, in recorded order. Empty for an unknown runID.
	RunSteps(ctx context.Context, runID string) ([]StepOccurrence, error)
}

// ErrPipelineNotFound is returned when a PipelineID resolves to no
// recorded run history.
var ErrPipelineNotFound = errors.New("pipeline not found")

// ErrRunNotFound is returned when a run id resolves to no recorded steps.
var ErrRunNotFound = errors.New("run not found")

// Service computes health, trends, step rankings, and usage from a Store.
type Service struct {
	store Store
}

// NewService returns a Service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// ListPipelines returns a page of tracked pipelines' summaries and health
// status, plus whether more pipelines beyond this page match filter.
func (s *Service) ListPipelines(ctx context.Context, window Window, filter PipelineListFilter) ([]Pipeline, bool, error) {
	refs, hasMore, err := s.store.ListPipelines(ctx, filter)
	if err != nil {
		return nil, false, fmt.Errorf("list pipelines: %w", err)
	}

	pipelines := make([]Pipeline, 0, len(refs))

	for _, ref := range refs {
		pipeline, err := s.summarize(ctx, ref, window)
		if err != nil {
			return nil, false, err
		}

		pipelines = append(pipelines, pipeline)
	}

	return pipelines, hasMore, nil
}

// GetPipeline returns one pipeline's health, duration trend, and
// failure-rate trend.
func (s *Service) GetPipeline(ctx context.Context, id PipelineID, window Window) (PipelineDetail, error) {
	ref, err := ParsePipelineID(id)
	if err != nil {
		return PipelineDetail{}, err
	}

	runs, err := s.store.PipelineRuns(ctx, ref, window.doubled())
	if err != nil {
		return PipelineDetail{}, fmt.Errorf("load pipeline runs: %w", err)
	}

	if len(runs) == 0 {
		return PipelineDetail{}, ErrPipelineNotFound
	}

	pipeline, err := s.summarizeFromRuns(ctx, ref, window, runs)
	if err != nil {
		return PipelineDetail{}, err
	}

	current, _ := splitWindow(runs, window)
	oldestFirst := reversed(current)

	return PipelineDetail{
		Pipeline:         pipeline,
		DurationTrend:    durationTrend(oldestFirst),
		FailureRateTrend: failureRateTrend(oldestFirst),
	}, nil
}

// GetPipelineSteps returns a pipeline's steps ranked by duration
// contribution, highest first.
func (s *Service) GetPipelineSteps(ctx context.Context, id PipelineID, window Window) ([]Step, error) {
	ref, err := ParsePipelineID(id)
	if err != nil {
		return nil, err
	}

	occurrences, err := s.store.PipelineSteps(ctx, ref, window)
	if err != nil {
		return nil, fmt.Errorf("load pipeline steps: %w", err)
	}

	return aggregateSteps(occurrences), nil
}

// ListFlakyRuns returns every run in which the named step failed within
// window, most-recent-first -- the step-scoped drill-down from a flaky
// step in GetPipelineSteps.
func (s *Service) ListFlakyRuns(ctx context.Context, id PipelineID, stepName string, window Window) ([]FlakyRun, error) {
	ref, err := ParsePipelineID(id)
	if err != nil {
		return nil, err
	}

	occurrences, err := s.store.PipelineSteps(ctx, ref, window)
	if err != nil {
		return nil, fmt.Errorf("load pipeline steps: %w", err)
	}

	var runs []FlakyRun

	for _, occ := range occurrences {
		if occ.Name != stepName || occ.Conclusion != "failure" {
			continue
		}

		runs = append(runs, FlakyRun{
			RunID:     occ.RunID,
			StartedAt: occ.RunStartedAt,
			ForgeURL:  occ.JobForgeURL,
		})
	}

	slices.SortFunc(runs, func(a, b FlakyRun) int {
		return compareStartedDesc(a.StartedAt, b.StartedAt)
	})

	return runs, nil
}

// GetRunSteps returns one run's own steps and statuses, in recorded
// order -- the drill-down a FlakyRun leads to.
func (s *Service) GetRunSteps(ctx context.Context, runID string) (RunDetail, error) {
	occurrences, err := s.store.RunSteps(ctx, runID)
	if err != nil {
		return RunDetail{}, fmt.Errorf("load run steps: %w", err)
	}

	if len(occurrences) == 0 {
		return RunDetail{}, ErrRunNotFound
	}

	steps := make([]RunStep, 0, len(occurrences))
	for _, occ := range occurrences {
		steps = append(steps, RunStep{
			Name:       occ.Name,
			Status:     occ.Status,
			Conclusion: occ.Conclusion,
			ForgeURL:   occ.JobForgeURL,
		})
	}

	return RunDetail{
		RunID:     runID,
		StartedAt: occurrences[0].RunStartedAt,
		Steps:     steps,
	}, nil
}

// compareStartedDesc orders most-recent-first, with an unknown (nil) time
// sorting last rather than panicking on the dereference.
func compareStartedDesc(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	default:
		return b.Compare(*a)
	}
}

// PipelineStepsGroup is one pipeline's flaky or failing steps, for the
// cross-pipeline unhealthy-steps overview.
type PipelineStepsGroup struct {
	PipelineID   PipelineID
	PipelineName string
	RepoID       string
	Steps        []Step
}

// ListUnhealthySteps returns every flaky or failing step across every
// tracked pipeline, grouped by pipeline. A pipeline with no such step
// contributes nothing to the result.
func (s *Service) ListUnhealthySteps(ctx context.Context, window Window) ([]PipelineStepsGroup, error) {
	refs, _, err := s.store.ListPipelines(ctx, PipelineListFilter{})
	if err != nil {
		return nil, fmt.Errorf("list pipelines: %w", err)
	}

	var groups []PipelineStepsGroup

	for _, ref := range refs {
		occurrences, err := s.store.PipelineSteps(ctx, ref, window)
		if err != nil {
			return nil, fmt.Errorf("load steps for %s: %w", ref.Name, err)
		}

		var unhealthy []Step
		for _, step := range aggregateSteps(occurrences) {
			if step.Flaky || step.FailureRate > 0 {
				unhealthy = append(unhealthy, step)
			}
		}

		if len(unhealthy) == 0 {
			continue
		}

		groups = append(groups, PipelineStepsGroup{
			PipelineID:   ref.ID(),
			PipelineName: ref.Name,
			RepoID:       ref.RepoID,
			Steps:        unhealthy,
		})
	}

	return groups, nil
}

// GetRepoUsage returns a repo's runner-minutes usage, broken down by
// workflow, highest first.
func (s *Service) GetRepoUsage(ctx context.Context, repoID string, window Window) ([]UsageEntry, error) {
	records, err := s.store.RepoUsage(ctx, repoID, window)
	if err != nil {
		return nil, fmt.Errorf("load repo usage: %w", err)
	}

	totals := map[string]float64{}

	var order []string

	for _, rec := range records {
		if _, ok := totals[rec.PipelineName]; !ok {
			order = append(order, rec.PipelineName)
		}

		totals[rec.PipelineName] += rec.ExecSeconds
	}

	entries := make([]UsageEntry, 0, len(order))
	for _, name := range order {
		entries = append(entries, UsageEntry{Workflow: name, RunnerMinutes: totals[name] / 60})
	}

	slices.SortFunc(entries, func(a, b UsageEntry) int {
		return cmp.Compare(b.RunnerMinutes, a.RunnerMinutes)
	})

	return entries, nil
}

func (s *Service) summarize(ctx context.Context, ref PipelineRef, window Window) (Pipeline, error) {
	runs, err := s.store.PipelineRuns(ctx, ref, window.doubled())
	if err != nil {
		return Pipeline{}, fmt.Errorf("load runs for %s: %w", ref.Name, err)
	}

	return s.summarizeFromRuns(ctx, ref, window, runs)
}

func (s *Service) summarizeFromRuns(ctx context.Context, ref PipelineRef, window Window, runs []RunRecord) (Pipeline, error) {
	current, prior := splitWindow(runs, window)

	occurrences, err := s.store.PipelineSteps(ctx, ref, window)
	if err != nil {
		return Pipeline{}, fmt.Errorf("load steps for %s: %w", ref.Name, err)
	}

	status, signals := computeHealth(current, prior, anyFlaky(aggregateSteps(occurrences)))

	return Pipeline{
		ID:               ref.ID(),
		RepoID:           ref.RepoID,
		Name:             ref.Name,
		HealthStatus:     status,
		TriggeredSignals: signals,
		LastRunAt:        lastRunAt(runs),
	}, nil
}

// lastRunAt returns the most recent run's start time, falling back to its
// completion time when it started before ingestion recorded a start (never
// observed in practice, but StartedAt is a pointer). runs is
// most-recent-first, per the Store contract.
func lastRunAt(runs []RunRecord) *time.Time {
	if len(runs) == 0 {
		return nil
	}

	if runs[0].StartedAt != nil {
		return runs[0].StartedAt
	}

	return runs[0].CompletedAt
}

func anyFlaky(steps []Step) bool {
	for _, s := range steps {
		if s.Flaky {
			return true
		}
	}

	return false
}

// splitWindow divides runsMostRecentFirst into the current window (the
// first window.RunCount runs) and the prior window (the next
// window.RunCount runs immediately before it).
func splitWindow(runsMostRecentFirst []RunRecord, window Window) (current, prior []RunRecord) {
	n := window.RunCount
	if n <= 0 {
		n = DefaultWindowRunCount
	}

	if len(runsMostRecentFirst) <= n {
		return runsMostRecentFirst, nil
	}

	current = runsMostRecentFirst[:n]

	end := min(2*n, len(runsMostRecentFirst))
	prior = runsMostRecentFirst[n:end]

	return current, prior
}

func reversed(runs []RunRecord) []RunRecord {
	out := make([]RunRecord, len(runs))
	for i, r := range runs {
		out[len(runs)-1-i] = r
	}

	return out
}

func computeHealth(currentRuns, priorRuns []RunRecord, hasFlakyStep bool) (HealthStatus, []Signal) {
	var signals []Signal

	if rate := overallFailureRate(currentRuns); rate >= unhealthyFailureRateThreshold {
		signals = append(signals, SignalFailureRate)
	}

	if regressed(currentRuns, priorRuns) {
		signals = append(signals, SignalDurationRegression)
	}

	if hasFlakyStep {
		signals = append(signals, SignalFlakyStep)
	}

	if len(signals) == 0 {
		return HealthHealthy, nil
	}

	return HealthUnhealthy, signals
}

func overallFailureRate(runs []RunRecord) float64 {
	var total, failed int

	for _, r := range runs {
		if r.Status != "completed" {
			continue
		}

		total++

		if r.Conclusion == "failure" {
			failed++
		}
	}

	return rate(failed, total)
}

func regressed(currentRuns, priorRuns []RunRecord) bool {
	currentP90 := runsP90Duration(currentRuns)
	priorP90 := runsP90Duration(priorRuns)

	if currentP90 == 0 || priorP90 == 0 {
		return false
	}

	return currentP90 >= priorP90*durationRegressionFactor
}

func runsP90Duration(runs []RunRecord) float64 {
	durations := make([]float64, 0, len(runs))

	for _, r := range runs {
		if secs, ok := durationSeconds(r.StartedAt, r.CompletedAt); ok {
			durations = append(durations, secs)
		}
	}

	if len(durations) == 0 {
		return 0
	}

	sort.Float64s(durations)

	return percentile(durations, 0.9)
}

// durationTrend buckets runsOldestFirst into consecutive groups and
// computes p50/p90 duration per bucket, so the trend has multiple points
// rather than a single aggregate.
func durationTrend(runsOldestFirst []RunRecord) Trend {
	var trend Trend

	for _, bucket := range bucketRuns(runsOldestFirst) {
		durations := make([]float64, 0, len(bucket))

		var ts time.Time

		for _, run := range bucket {
			if secs, ok := durationSeconds(run.StartedAt, run.CompletedAt); ok {
				durations = append(durations, secs)
			}

			if run.StartedAt != nil {
				ts = *run.StartedAt
			}
		}

		if len(durations) == 0 {
			continue
		}

		sort.Float64s(durations)
		trend.Timestamps = append(trend.Timestamps, ts)
		trend.P50 = append(trend.P50, percentile(durations, 0.5))
		trend.P90 = append(trend.P90, percentile(durations, 0.9))
	}

	return trend
}

func failureRateTrend(runsOldestFirst []RunRecord) Trend {
	var trend Trend

	for _, bucket := range bucketRuns(runsOldestFirst) {
		var total, failed int

		var ts time.Time

		for _, run := range bucket {
			if run.Status != "completed" {
				continue
			}

			total++
			if run.Conclusion == "failure" {
				failed++
			}

			if run.StartedAt != nil {
				ts = *run.StartedAt
			}
		}

		if total == 0 {
			continue
		}

		trend.Timestamps = append(trend.Timestamps, ts)
		trend.Rate = append(trend.Rate, rate(failed, total))
	}

	return trend
}

func bucketRuns(runsOldestFirst []RunRecord) [][]RunRecord {
	var buckets [][]RunRecord

	for i := 0; i < len(runsOldestFirst); i += trendBucketSize {
		end := min(i+trendBucketSize, len(runsOldestFirst))
		buckets = append(buckets, runsOldestFirst[i:end])
	}

	return buckets
}

// stepAcc accumulates every recorded occurrence of one named step.
type stepAcc struct {
	execSecs   []float64
	queueSecs  []float64
	total      int
	failed     int
	sawSuccess bool
	sawFailure bool
	forgeURL   string
}

func (a *stepAcc) add(occ StepOccurrence) {
	if secs, ok := durationSeconds(occ.StartedAt, occ.CompletedAt); ok {
		a.execSecs = append(a.execSecs, secs)
	}

	if secs, ok := durationSeconds(occ.JobQueuedAt, occ.JobStartedAt); ok {
		a.queueSecs = append(a.queueSecs, secs)
	}

	if occ.Status == "completed" {
		a.total++

		switch occ.Conclusion {
		case "success":
			a.sawSuccess = true
		case "failure":
			a.failed++

			a.sawFailure = true
		}
	}

	if a.forgeURL == "" {
		a.forgeURL = occ.JobForgeURL
	}
}

// aggregateSteps groups occurrences by step name and computes each one's
// duration contribution, queue/exec split, failure rate, and flakiness.
// Steps are returned ranked by duration contribution, highest first.
func aggregateSteps(occurrences []StepOccurrence) []Step {
	byName := map[string]*stepAcc{}

	var order []string

	for _, occ := range occurrences {
		a, ok := byName[occ.Name]
		if !ok {
			a = &stepAcc{}
			byName[occ.Name] = a

			order = append(order, occ.Name)
		}

		a.add(occ)
	}

	steps := make([]Step, 0, len(order))

	for _, name := range order {
		a := byName[name]

		steps = append(steps, Step{
			ID:                          name,
			Name:                        name,
			DurationContributionSeconds: average(a.execSecs),
			QueueSeconds:                average(a.queueSecs),
			ExecSeconds:                 average(a.execSecs),
			FailureRate:                 rate(a.failed, a.total),
			FailureCount:                a.failed,
			Flaky:                       a.sawSuccess && a.sawFailure,
			ForgeURL:                    a.forgeURL,
		})
	}

	slices.SortFunc(steps, func(a, b Step) int {
		return cmp.Compare(b.DurationContributionSeconds, a.DurationContributionSeconds)
	})

	return steps
}

func durationSeconds(start, end *time.Time) (float64, bool) {
	if start == nil || end == nil {
		return 0, false
	}

	return end.Sub(*start).Seconds(), true
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

func rate(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}

	return float64(numerator) / float64(denominator)
}

// percentile returns the p-th percentile (0-1) of sorted (ascending),
// using the nearest-rank method.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	rank := int(math.Ceil(p*float64(len(sorted)))) - 1
	rank = max(rank, 0)
	rank = min(rank, len(sorted)-1)

	return sorted[rank]
}

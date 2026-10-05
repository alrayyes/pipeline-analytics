package metrics

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"
)

// InsightWindow is a trailing span of time, for the cross-pipeline failure
// insights. It's separate from Window, the per-pipeline trailing run count:
// "failures this week" has no meaning as a number of runs, because a quiet
// pipeline and a busy one would cover wildly different spans.
type InsightWindow struct {
	Duration time.Duration
}

// DefaultInsightWindow is used when no window, or an unrecognised one, is
// requested.
const DefaultInsightWindow = 7 * 24 * time.Hour

// ParseInsightWindow interprets the API's window parameter for the insights
// endpoint: 24h, 7d or 30d. Anything else, including a bare run count, falls
// back to DefaultInsightWindow.
func ParseInsightWindow(raw string) InsightWindow {
	switch raw {
	case "24h":
		return InsightWindow{Duration: 24 * time.Hour}
	case "7d":
		return InsightWindow{Duration: 7 * 24 * time.Hour}
	case "30d":
		return InsightWindow{Duration: 30 * 24 * time.Hour}
	default:
		return InsightWindow{Duration: DefaultInsightWindow}
	}
}

// Label is the window as the API names it ("24h", "7d", "30d"). A duration
// the parser would never produce labels as the default, so a response always
// names a window a client can send back.
func (w InsightWindow) Label() string {
	switch w.Duration {
	case 24 * time.Hour:
		return "24h"
	case 30 * 24 * time.Hour:
		return "30d"
	default:
		return "7d"
	}
}

// MaxBranchLength bounds a branch filter (the OpenAPI spec's BranchFilter
// maxLength).
const MaxBranchLength = 255

// InsightFilter scopes the insights to one repo and/or forge. The zero value
// covers every tracked repo.
type InsightFilter struct {
	RepoID string
	Forge  string
	// Branch, when set, covers only runs on that branch.
	Branch string
}

// RunWindowFilter selects runs that started in [Since, Until), optionally
// scoped to one repo and/or forge.
type RunWindowFilter struct {
	RepoID string
	Forge  string
	Branch string
	Since  time.Time
	Until  time.Time
}

// BranchCount is a branch and how many runs on it started in a window.
type BranchCount struct {
	Name     string
	RunCount int
}

// WindowRun is a run with the pipeline it belongs to.
type WindowRun struct {
	Pipeline PipelineRef
	Run      RunRecord
}

// WindowStep is one recorded step execution with the pipeline it ran in.
// Step.RunStartedAt is the owning run's start, which is what a window
// bounds.
type WindowStep struct {
	Pipeline PipelineRef
	Step     StepOccurrence
}

// StageFailureCount is how many times one step failed in the window. "Stage"
// means step: forges have no stage taxonomy.
type StageFailureCount struct {
	Step     string
	Failures int
	// Share is Failures over every failed-step occurrence in the window.
	Share float64
}

// CategoryCount is how many failed-step occurrences fall in one failure
// category, and that category's share of all of them.
type CategoryCount struct {
	Category    FailureCategory
	Occurrences int
	Share       float64
}

// FailureGroup is one step's failures across every pipeline it failed in.
type FailureGroup struct {
	Step string
	// Category is a heuristic from Conclusion and the step's name, never from
	// a log; see CategorizeFailure.
	Category FailureCategory
	// Conclusion is the group's most common conclusion, shown so a heuristic
	// category can be checked against it.
	Conclusion  string
	Occurrences int
	Pipelines   []PipelineRef
}

// FailingPipeline is one pipeline's run and failed-run counts in the window.
type FailingPipeline struct {
	Pipeline   PipelineRef
	Runs       int
	FailedRuns int
}

// FailureInsights is the failure overview for one window.
type FailureInsights struct {
	// Window is the window these figures cover: the one requested, or the
	// default when the request named none or an unknown one.
	Window     InsightWindow
	TotalRuns  int
	FailedRuns int
	// PassRate is the fraction in [0, 1] of concluded runs that succeeded.
	// Nil when no run concluded, since "no data" isn't 0%.
	PassRate *float64
	// PassRateDelta is the change in percentage points against the preceding
	// window of equal length. Nil when either window has no concluded run.
	PassRateDelta *float64
	// MTTR is the mean time to recovery: for each pipeline, from the
	// completion of a failed run to the completion of the next successful
	// one, averaged over every recovery in the window. Nil when nothing
	// recovered, since an unrecovered failure has no recovery time.
	MTTR                *time.Duration
	TopFailingPipelines []FailingPipeline
	// FlakyStepRatio is flaky steps over distinct steps, counting a step per
	// pipeline. A step is flaky when it both passed and failed in the window.
	FlakyStepRatio    float64
	StageDistribution []StageFailureCount
	// CategoryBreakdown is the failed-step occurrences by failure category,
	// heaviest first; empty when nothing failed.
	CategoryBreakdown []CategoryCount
	FailureGroups     []FailureGroup
}

// GetFailureInsights computes the failure overview for the window ending at
// now. now is a parameter, not read from the clock, so the preceding-window
// comparison is testable.
func (s *Service) GetFailureInsights(ctx context.Context, now time.Time, window InsightWindow, filter InsightFilter) (FailureInsights, error) {
	boundary := now.Add(-window.Duration)

	runs, err := s.store.WindowRuns(ctx, RunWindowFilter{
		RepoID: filter.RepoID,
		Forge:  filter.Forge,
		Branch: filter.Branch,
		Since:  boundary.Add(-window.Duration),
		Until:  now,
	})
	if err != nil {
		return FailureInsights{}, fmt.Errorf("list window runs: %w", err)
	}

	var current, prior []WindowRun

	for _, wr := range runs {
		if wr.Run.StartedAt == nil {
			continue
		}

		if wr.Run.StartedAt.Before(boundary) {
			prior = append(prior, wr)
		} else {
			current = append(current, wr)
		}
	}

	steps, err := s.store.WindowSteps(ctx, RunWindowFilter{
		RepoID: filter.RepoID,
		Forge:  filter.Forge,
		Branch: filter.Branch,
		Since:  boundary,
		Until:  now,
	})
	if err != nil {
		return FailureInsights{}, fmt.Errorf("list window steps: %w", err)
	}

	currentRate, currentFailed := passRate(current)
	priorRate, _ := passRate(prior)

	insights := FailureInsights{
		Window:              window,
		TotalRuns:           len(current),
		FailedRuns:          currentFailed,
		PassRate:            currentRate,
		MTTR:                meanTimeToRecovery(current),
		TopFailingPipelines: topFailingPipelines(current),
		FlakyStepRatio:      flakyStepRatio(steps),
		StageDistribution:   stageDistribution(steps),
		CategoryBreakdown:   categoryBreakdown(steps),
		FailureGroups:       failureGroups(steps),
	}

	if currentRate != nil && priorRate != nil {
		delta := (*currentRate - *priorRate) * 100
		insights.PassRateDelta = &delta
	}

	return insights, nil
}

// isFailed reports whether a run concluded in a failure. A timeout is a
// failure here; a cancelled or skipped run is neither one nor a pass.
func isFailed(conclusion string) bool {
	return conclusion == "failure" || conclusion == "timed_out"
}

// passRate returns the fraction of concluded runs that succeeded (nil if
// none concluded) and the failed count.
func passRate(runs []WindowRun) (rate *float64, failed int) {
	var passed, concluded int

	for _, wr := range runs {
		switch {
		case wr.Run.Conclusion == "success":
			passed++
			concluded++
		case isFailed(wr.Run.Conclusion):
			failed++
			concluded++
		}
	}

	if concluded == 0 {
		return nil, failed
	}

	r := float64(passed) / float64(concluded)

	return &r, failed
}

func topFailingPipelines(runs []WindowRun) []FailingPipeline {
	byPipeline := make(map[PipelineRef]*FailingPipeline)

	for _, wr := range runs {
		fp, ok := byPipeline[wr.Pipeline]
		if !ok {
			fp = &FailingPipeline{Pipeline: wr.Pipeline}
			byPipeline[wr.Pipeline] = fp
		}

		fp.Runs++

		if isFailed(wr.Run.Conclusion) {
			fp.FailedRuns++
		}
	}

	var failing []FailingPipeline

	for _, fp := range byPipeline {
		if fp.FailedRuns > 0 {
			failing = append(failing, *fp)
		}
	}

	slices.SortFunc(failing, func(a, b FailingPipeline) int {
		return cmp.Or(
			cmp.Compare(b.FailedRuns, a.FailedRuns),
			cmp.Compare(a.Pipeline.RepoID, b.Pipeline.RepoID),
			cmp.Compare(a.Pipeline.Name, b.Pipeline.Name),
		)
	})

	return failing
}

// meanTimeToRecovery averages, over every recovery, the time from a failed
// run's completion to the next successful run's completion in the same
// pipeline. A run of consecutive failures counts from the first; cancelled
// and unfinished runs neither start nor end an outage.
func meanTimeToRecovery(runs []WindowRun) *time.Duration {
	byPipeline := make(map[PipelineRef][]WindowRun)

	for _, wr := range runs {
		if wr.Run.Conclusion == "success" || isFailed(wr.Run.Conclusion) {
			byPipeline[wr.Pipeline] = append(byPipeline[wr.Pipeline], wr)
		}
	}

	var (
		total      time.Duration
		recoveries int
	)

	for _, pipelineRuns := range byPipeline {
		slices.SortFunc(pipelineRuns, func(a, b WindowRun) int {
			return finishedAt(a.Run).Compare(finishedAt(b.Run))
		})

		var downSince *time.Time

		for _, wr := range pipelineRuns {
			at := finishedAt(wr.Run)

			switch {
			case isFailed(wr.Run.Conclusion) && downSince == nil:
				downSince = &at
			case wr.Run.Conclusion == "success" && downSince != nil:
				total += at.Sub(*downSince)
				recoveries++
				downSince = nil
			}
		}
	}

	if recoveries == 0 {
		return nil
	}

	mean := total / time.Duration(recoveries)

	return &mean
}

// finishedAt is when a run is known to have ended, falling back to its start
// when no completion was recorded.
func finishedAt(run RunRecord) time.Time {
	if run.CompletedAt != nil {
		return *run.CompletedAt
	}

	if run.StartedAt != nil {
		return *run.StartedAt
	}

	return time.Time{}
}

// isFailedStep reports whether a step execution ended in a failure. As for
// runs, a timeout is a failure; a skipped or cancelled step is not.
func isFailedStep(occ StepOccurrence) bool {
	return occ.Status == "completed" && isFailed(occ.Conclusion)
}

func stageDistribution(steps []WindowStep) []StageFailureCount {
	counts := make(map[string]int)

	for _, ws := range steps {
		if isFailedStep(ws.Step) {
			counts[ws.Step.Name]++
		}
	}

	var total int
	for _, failures := range counts {
		total += failures
	}

	distribution := make([]StageFailureCount, 0, len(counts))
	for name, failures := range counts {
		distribution = append(distribution, StageFailureCount{Step: name, Failures: failures, Share: rate(failures, total)})
	}

	slices.SortFunc(distribution, func(a, b StageFailureCount) int {
		return cmp.Or(cmp.Compare(b.Failures, a.Failures), cmp.Compare(a.Step, b.Step))
	})

	return distribution
}

func failureGroups(steps []WindowStep) []FailureGroup {
	type acc struct {
		occurrences int
		conclusions map[string]int
		pipelines   map[PipelineRef]struct{}
	}

	byStep := make(map[string]*acc)

	for _, ws := range steps {
		if !isFailedStep(ws.Step) {
			continue
		}

		a, ok := byStep[ws.Step.Name]
		if !ok {
			a = &acc{conclusions: map[string]int{}, pipelines: map[PipelineRef]struct{}{}}
			byStep[ws.Step.Name] = a
		}

		a.occurrences++
		a.conclusions[ws.Step.Conclusion]++
		a.pipelines[ws.Pipeline] = struct{}{}
	}

	groups := make([]FailureGroup, 0, len(byStep))

	for name, a := range byStep {
		conclusion := mostCommon(a.conclusions)

		pipelines := make([]PipelineRef, 0, len(a.pipelines))
		for ref := range a.pipelines {
			pipelines = append(pipelines, ref)
		}

		slices.SortFunc(pipelines, func(x, y PipelineRef) int {
			return cmp.Or(cmp.Compare(x.RepoID, y.RepoID), cmp.Compare(x.Name, y.Name))
		})

		groups = append(groups, FailureGroup{
			Step:        name,
			Category:    CategorizeFailure(name, conclusion),
			Conclusion:  conclusion,
			Occurrences: a.occurrences,
			Pipelines:   pipelines,
		})
	}

	slices.SortFunc(groups, func(a, b FailureGroup) int {
		return cmp.Or(cmp.Compare(b.Occurrences, a.Occurrences), cmp.Compare(a.Step, b.Step))
	})

	return groups
}

// mostCommon returns the key with the highest count, the alphabetically
// first on a tie so the result doesn't depend on map order.
func mostCommon(counts map[string]int) string {
	var (
		best      string
		bestCount int
	)

	for key, count := range counts {
		if count > bestCount || (count == bestCount && key < best) {
			best, bestCount = key, count
		}
	}

	return best
}

// flakyStepRatio is flaky steps over distinct steps, a step being one name
// within one pipeline: "test" flaking in web says nothing about "test" in
// api. Flakiness is the existing definition (it both passed and failed), via
// aggregateSteps.
func flakyStepRatio(steps []WindowStep) float64 {
	byPipeline := make(map[PipelineRef][]StepOccurrence)
	for _, ws := range steps {
		byPipeline[ws.Pipeline] = append(byPipeline[ws.Pipeline], ws.Step)
	}

	var total, flaky int

	for _, occurrences := range byPipeline {
		for _, step := range aggregateSteps(occurrences) {
			total++

			if step.Flaky {
				flaky++
			}
		}
	}

	return rate(flaky, total)
}

// categoryBreakdown counts failed-step occurrences by failure category, heaviest
// first (equal weights by category name, so the order is stable). A category
// is decided per step name and its most common conclusion, the same as the
// failure groups, so the two views can't disagree about which category a
// step is in.
func categoryBreakdown(steps []WindowStep) []CategoryCount {
	var total int

	counts := make(map[FailureCategory]int)

	for _, group := range failureGroups(steps) {
		counts[group.Category] += group.Occurrences
		total += group.Occurrences
	}

	breakdown := make([]CategoryCount, 0, len(counts))
	for category, occurrences := range counts {
		breakdown = append(breakdown, CategoryCount{Category: category, Occurrences: occurrences, Share: rate(occurrences, total)})
	}

	slices.SortFunc(breakdown, func(a, b CategoryCount) int {
		return cmp.Or(cmp.Compare(b.Occurrences, a.Occurrences), cmp.Compare(a.Category, b.Category))
	})

	return breakdown
}

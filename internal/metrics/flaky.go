package metrics

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"
)

// FlakyMatrixRuns is how many of a flaky step's most recent runs its result
// matrix shows.
const FlakyMatrixRuns = 40

// FlakyStep is one step of one pipeline that both passed and failed in the
// window, with what the flaky view shows for it.
type FlakyStep struct {
	Pipeline PipelineRef
	Name     string
	// FlakeRate is the share of the step's completed runs in the window that
	// failed, the same figure Step.FailureRate reports.
	FlakeRate float64
	// RunCount is the step's completed runs in the window, not just the
	// ones in RecentOutcomes.
	RunCount int
	// RecentOutcomes is the result in the step's most recent runs (at most
	// FlakyMatrixRuns), oldest first.
	RecentOutcomes []Outcome
}

// ListFlakySteps returns the flaky steps across every pipeline in the window
// ending at now, ranked by flake rate, then run count, then pipeline and step
// name. limit and offset page the ranked list; a limit of zero or less
// returns everything, and hasMore says whether entries follow the page.
//
// A step is flaky by the same definition the step lists use (aggregateSteps):
// it completed with both a success and a failure in the window, and a step
// is one name within one pipeline.
func (s *Service) ListFlakySteps(ctx context.Context, now time.Time, window InsightWindow, filter InsightFilter, limit, offset int) ([]FlakyStep, bool, error) {
	steps, err := s.store.WindowSteps(ctx, RunWindowFilter{
		RepoID: filter.RepoID,
		Forge:  filter.Forge,
		Since:  now.Add(-window.Duration),
		Until:  now,
	})
	if err != nil {
		return nil, false, fmt.Errorf("list window steps: %w", err)
	}

	type key struct {
		pipeline PipelineRef
		name     string
	}

	byStep := make(map[key][]StepOccurrence)

	for _, ws := range steps {
		if ws.Step.Status != "completed" {
			continue
		}

		k := key{pipeline: ws.Pipeline, name: ws.Step.Name}
		byStep[k] = append(byStep[k], ws.Step)
	}

	var flaky []FlakyStep

	for k, occurrences := range byStep {
		if entry, ok := flakyStep(k.pipeline, k.name, occurrences); ok {
			flaky = append(flaky, entry)
		}
	}

	slices.SortFunc(flaky, func(a, b FlakyStep) int {
		return cmp.Or(
			cmp.Compare(b.FlakeRate, a.FlakeRate),
			cmp.Compare(b.RunCount, a.RunCount),
			cmp.Compare(a.Pipeline.RepoID, b.Pipeline.RepoID),
			cmp.Compare(a.Pipeline.Name, b.Pipeline.Name),
			cmp.Compare(a.Name, b.Name),
		)
	})

	page, hasMore := paginate(flaky, limit, offset)

	return page, hasMore, nil
}

// flakyStep builds the entry for one step's completed occurrences, or reports
// false when it didn't both pass and fail.
func flakyStep(pipeline PipelineRef, name string, occurrences []StepOccurrence) (FlakyStep, bool) {
	slices.SortStableFunc(occurrences, func(a, b StepOccurrence) int {
		return startedAtCompare(a.RunStartedAt, b.RunStartedAt)
	})

	var passed, failed int

	for _, occ := range occurrences {
		switch occ.Conclusion {
		case "success":
			passed++
		case "failure":
			failed++
		}
	}

	if passed == 0 || failed == 0 {
		return FlakyStep{}, false
	}

	recent := occurrences
	if len(recent) > FlakyMatrixRuns {
		recent = recent[len(recent)-FlakyMatrixRuns:]
	}

	outcomes := make([]Outcome, 0, len(recent))
	for _, occ := range recent {
		outcomes = append(outcomes, OutcomeOf(occ.Status, occ.Conclusion))
	}

	return FlakyStep{
		Pipeline:       pipeline,
		Name:           name,
		FlakeRate:      rate(failed, len(occurrences)),
		RunCount:       len(occurrences),
		RecentOutcomes: outcomes,
	}, true
}

// startedAtCompare orders oldest first, runs with no start time before any
// that have one.
func startedAtCompare(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	default:
		return a.Compare(*b)
	}
}

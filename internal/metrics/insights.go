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

// InsightFilter scopes the insights to one repo and/or forge. The zero value
// covers every tracked repo.
type InsightFilter struct {
	RepoID string
	Forge  string
}

// RunWindowFilter selects runs that started in [Since, Until), optionally
// scoped to one repo and/or forge.
type RunWindowFilter struct {
	RepoID string
	Forge  string
	Since  time.Time
	Until  time.Time
}

// WindowRun is a run with the pipeline it belongs to.
type WindowRun struct {
	Pipeline PipelineRef
	Run      RunRecord
}

// FailingPipeline is one pipeline's run and failed-run counts in the window.
type FailingPipeline struct {
	Pipeline   PipelineRef
	Runs       int
	FailedRuns int
}

// FailureInsights is the failure overview for one window.
type FailureInsights struct {
	TotalRuns  int
	FailedRuns int
	// PassRate is the fraction in [0, 1] of concluded runs that succeeded.
	// Nil when no run concluded, since "no data" isn't 0%.
	PassRate *float64
	// PassRateDelta is the change in percentage points against the preceding
	// window of equal length. Nil when either window has no concluded run.
	PassRateDelta       *float64
	TopFailingPipelines []FailingPipeline
}

// GetFailureInsights computes the failure overview for the window ending at
// now. now is a parameter, not read from the clock, so the preceding-window
// comparison is testable.
func (s *Service) GetFailureInsights(ctx context.Context, now time.Time, window InsightWindow, filter InsightFilter) (FailureInsights, error) {
	boundary := now.Add(-window.Duration)

	runs, err := s.store.WindowRuns(ctx, RunWindowFilter{
		RepoID: filter.RepoID,
		Forge:  filter.Forge,
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

	currentRate, currentFailed := passRate(current)
	priorRate, _ := passRate(prior)

	insights := FailureInsights{
		TotalRuns:           len(current),
		FailedRuns:          currentFailed,
		PassRate:            currentRate,
		TopFailingPipelines: topFailingPipelines(current),
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

package metrics

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

// PipelineSort is the order a pipeline list comes back in.
type PipelineSort string

// The pipeline list orders. SortByName is by repository then name, the order
// the store returns; SortByLastRun is the most recently run first.
const (
	SortByName    PipelineSort = "name"
	SortByLastRun PipelineSort = "lastRun"
)

// ErrInvalidHealthFilter and ErrInvalidPipelineSort are returned by the
// parsers for a value outside the documented set, so a typo can't read as
// "no unhealthy pipelines".
var (
	ErrInvalidHealthFilter = errors.New("invalid health filter")
	ErrInvalidPipelineSort = errors.New("invalid pipeline sort")
)

// ParseHealthFilter interprets the API's health query parameter. Empty means
// no filter.
func ParseHealthFilter(raw string) (HealthStatus, error) {
	switch status := HealthStatus(raw); status {
	case "", HealthHealthy, HealthUnhealthy:
		return status, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidHealthFilter, raw)
	}
}

// ParsePipelineSort interprets the API's sort query parameter. Empty means
// the default, SortByName.
func ParsePipelineSort(raw string) (PipelineSort, error) {
	switch order := PipelineSort(raw); order {
	case "":
		return SortByName, nil
	case SortByName, SortByLastRun:
		return order, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidPipelineSort, raw)
	}
}

// listAcrossPipelines serves a list that needs every matching pipeline
// summarised before it can be cut into a page: a health filter, because
// health is computed from runs and steps rather than stored, and a last-run
// sort, which orders the whole set. It summarises, filters and sorts the full
// match, then pages the result, so page sizes and hasMore are right.
func (s *Service) listAcrossPipelines(ctx context.Context, window Window, filter PipelineListFilter) ([]Pipeline, bool, error) {
	all := filter
	all.Limit, all.Offset = 0, 0

	refs, _, err := s.store.ListPipelines(ctx, all)
	if err != nil {
		return nil, false, fmt.Errorf("list pipelines: %w", err)
	}

	pipelines := make([]Pipeline, 0, len(refs))

	for _, ref := range refs {
		pipeline, err := s.summarize(ctx, ref, window)
		if err != nil {
			return nil, false, err
		}

		if filter.Health != "" && pipeline.HealthStatus != filter.Health {
			continue
		}

		pipelines = append(pipelines, pipeline)
	}

	if filter.Sort == SortByLastRun {
		// The store's order (repository, then name) is the tie-break, so the
		// stable sort keeps the result deterministic.
		slices.SortStableFunc(pipelines, func(a, b Pipeline) int {
			return compareLastRunDesc(a.LastRunAt, b.LastRunAt)
		})
	}

	page, hasMore := paginate(pipelines, filter.Limit, filter.Offset)

	return page, hasMore, nil
}

// compareLastRunDesc orders a later run before an earlier one and a pipeline
// with no runs after every pipeline that has some.
func compareLastRunDesc(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	default:
		return cmp.Compare(b.UnixNano(), a.UnixNano())
	}
}

// paginate cuts a page out of list. A zero limit means everything from
// offset, and hasMore is true only when matches remain beyond the page.
func paginate(list []Pipeline, limit, offset int) ([]Pipeline, bool) {
	offset = min(max(offset, 0), len(list))

	if limit <= 0 {
		return list[offset:], false
	}

	end := min(offset+limit, len(list))

	return list[offset:end], len(list) > end
}

package metrics

import (
	"context"
	"fmt"
	"time"
)

// ListBranches returns the branches with runs that started in the trailing
// window, busiest first. now is a parameter so the window is testable.
func (s *Service) ListBranches(ctx context.Context, now time.Time, window InsightWindow, filter InsightFilter) ([]BranchCount, error) {
	branches, err := s.store.WindowBranches(ctx, RunWindowFilter{
		RepoID: filter.RepoID,
		Forge:  filter.Forge,
		Since:  now.Add(-window.Duration),
		Until:  now,
	})
	if err != nil {
		return nil, fmt.Errorf("list window branches: %w", err)
	}

	return branches, nil
}

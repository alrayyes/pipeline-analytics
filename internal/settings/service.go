package settings

import (
	"context"
	"fmt"
	"slices"
)

// Service resolves stored overrides against documented defaults, and
// validates an update before it ever reaches Store.
type Service struct {
	store Store
}

// NewService returns a Service backed by store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Get returns userID's settings, every field resolved against its
// documented default.
func (s *Service) Get(ctx context.Context, userID string) (Settings, error) {
	raw, err := s.store.Get(ctx, userID)
	if err != nil {
		return Settings{}, fmt.Errorf("get settings: %w", err)
	}

	return resolve(raw), nil
}

// Update validates every key/value in updates, then applies all of them in
// one Patch call if -- and only if -- every one is valid; an invalid key or
// value rejects the whole request, leaving every previously stored value
// unchanged (account-settings/spec.md's "Invalid value is rejected").
func (s *Service) Update(ctx context.Context, userID string, updates map[string]*string) (Settings, error) {
	for key, value := range updates {
		if !slices.Contains(Keys, key) {
			return Settings{}, fmt.Errorf("%w: %q", ErrInvalidKey, key)
		}

		if value == nil {
			continue
		}

		if allowed, hasEnum := enumValues[key]; hasEnum && !slices.Contains(allowed, *value) {
			return Settings{}, fmt.Errorf("%w: %q for %q", ErrInvalidValue, *value, key)
		}
	}

	raw, err := s.store.Patch(ctx, userID, updates)
	if err != nil {
		return Settings{}, fmt.Errorf("patch settings: %w", err)
	}

	return resolve(raw), nil
}

// ResetPipelinesFilters clears the three Pipelines-page filter settings
// back to their defaults in one request, leaving every other setting
// (theme, forge filter) untouched -- account-settings/spec.md's "Reset
// Pipelines filters to defaults".
func (s *Service) ResetPipelinesFilters(ctx context.Context, userID string) (Settings, error) {
	updates := make(map[string]*string, len(PipelinesFilterKeys))
	for _, key := range PipelinesFilterKeys {
		updates[key] = nil
	}

	return s.Update(ctx, userID, updates)
}

func resolve(raw map[string]string) Settings {
	return Settings{
		Theme:                 valueOr(raw, KeyTheme, DefaultTheme),
		ForgeFilter:           valueOr(raw, KeyForgeFilter, DefaultForgeFilter),
		PipelinesHealthFilter: valueOr(raw, KeyPipelinesHealthFilter, DefaultPipelinesHealthFilter),
		PipelinesRepoSelector: valueOr(raw, KeyPipelinesRepoSelector, DefaultPipelinesRepoSelector),
		PipelinesSortOrder:    valueOr(raw, KeyPipelinesSortOrder, DefaultPipelinesSortOrder),
	}
}

func valueOr(raw map[string]string, key, def string) string {
	if v, ok := raw[key]; ok {
		return v
	}

	return def
}

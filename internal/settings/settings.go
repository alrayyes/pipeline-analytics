// Package settings holds the account-settings domain: the dashboard
// account's persisted UI preferences (theme and the Pipelines/Repos page
// filters). See
// openspec/changes/persist-account-settings/specs/account-settings/spec.md.
package settings

import (
	"context"
	"errors"
)

// The setting keys, as stored in the JSON blob and used as PATCH request
// body keys -- see Store's doc comment for why a raw map keyed by these
// rather than a struct with pointer fields.
const (
	KeyTheme                 = "theme"
	KeyForgeFilter           = "forgeFilter"
	KeyPipelinesHealthFilter = "pipelinesHealthFilter"
	KeyPipelinesRepoSelector = "pipelinesRepoSelector"
	KeyPipelinesSortOrder    = "pipelinesSortOrder"
	KeyTelemetryWindow       = "telemetryWindow"
)

// Keys lists every known setting key, in a stable order -- for validating a
// PATCH request's keys and for the "reset Pipelines filters" convenience
// that clears exactly the Pipelines-prefixed ones.
var Keys = []string{
	KeyTheme,
	KeyForgeFilter,
	KeyPipelinesHealthFilter,
	KeyPipelinesRepoSelector,
	KeyPipelinesSortOrder,
	KeyTelemetryWindow,
}

// PipelinesFilterKeys are the three settings "reset filters" clears in one
// request -- account-settings/spec.md's "Reset Pipelines filters to
// defaults".
var PipelinesFilterKeys = []string{
	KeyPipelinesHealthFilter,
	KeyPipelinesRepoSelector,
	KeyPipelinesSortOrder,
}

// The documented default for each setting, resolved server-side whenever a
// key has never been explicitly set for the account (account-settings/
// spec.md's "Settings defaults"). PipelinesRepoSelector's default ("all")
// is a sentinel meaning "every tracked repo", not a real repo ID.
const (
	DefaultTheme                 = "system"
	DefaultForgeFilter           = "all"
	DefaultPipelinesHealthFilter = "unhealthy"
	DefaultPipelinesRepoSelector = "all"
	DefaultPipelinesSortOrder    = "name"
	DefaultTelemetryWindow       = "7d"
)

// enumValues holds the valid values for every setting that has a closed
// set -- PipelinesRepoSelector is deliberately absent: it's a repo ID (or
// "all"), which this package has no way to validate against the set of
// currently-tracked repos without a store dependency this domain doesn't
// otherwise need.
var enumValues = map[string][]string{
	KeyTheme:                 {"light", "dark", "system"},
	KeyForgeFilter:           {"all", "github", "forgejo"},
	KeyPipelinesHealthFilter: {"all", "healthy", "unhealthy"},
	KeyPipelinesSortOrder:    {"name", "lastRun"},
	KeyTelemetryWindow:       {"24h", "7d", "30d"},
}

// MaxRepoSelectorLength bounds the one free-text setting, a repo id or "all",
// so a caller can't store an arbitrarily large value (the OpenAPI spec's
// SettingsUpdate.pipelinesRepoSelector maxLength).
const MaxRepoSelectorLength = 128

// ErrInvalidKey is returned by Update for a key outside Keys.
var ErrInvalidKey = errors.New("unrecognized setting key")

// ErrInvalidValue is returned by Update for a value outside its setting's
// documented set (account-settings/spec.md's "Invalid value is rejected").
var ErrInvalidValue = errors.New("invalid setting value")

// Settings is every setting, fully resolved: a stored override merged onto
// its documented default, so a field here is never ambiguous between
// "unset" and "empty".
type Settings struct {
	Theme                 string
	ForgeFilter           string
	PipelinesHealthFilter string
	PipelinesRepoSelector string
	PipelinesSortOrder    string
	TelemetryWindow       string
}

// Defaults returns the documented default for every setting, the values a
// cleared key resolves to.
func Defaults() Settings {
	return Settings{
		Theme:                 DefaultTheme,
		ForgeFilter:           DefaultForgeFilter,
		PipelinesHealthFilter: DefaultPipelinesHealthFilter,
		PipelinesRepoSelector: DefaultPipelinesRepoSelector,
		PipelinesSortOrder:    DefaultPipelinesSortOrder,
		TelemetryWindow:       DefaultTelemetryWindow,
	}
}

// Store is the port account settings are persisted through. It works in
// terms of the raw, possibly-partial override map (only explicitly-set
// keys present) rather than a resolved Settings struct, so it -- and the
// wire format -- have no opinion on defaults; Service owns default
// resolution.
type Store interface {
	// Get returns userID's stored overrides. A key absent from the result
	// has never been set (or was cleared) and should fall back to its
	// default. A user with no settings row yet returns an empty map, not
	// an error.
	Get(ctx context.Context, userID string) (map[string]string, error)
	// Patch applies updates to userID's stored overrides: a nil value
	// clears that key (reverting it to its default on the next Get), a
	// non-nil value sets it. Returns the full resulting override map.
	Patch(ctx context.Context, userID string, updates map[string]*string) (map[string]string, error)
}

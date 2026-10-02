package metrics_test

import (
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestCategorizeFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		step       string
		conclusion string
		want       metrics.FailureCategory
	}{
		{"a timed-out step is a timeout whatever it's called", "go test", "timed_out", metrics.CategoryNetworkTimeouts},
		{"a test step", "Run unit tests", "failure", metrics.CategoryCodeTests},
		{"a lint step", "golangci-lint", "failure", metrics.CategoryCodeTests},
		{"an end-to-end step", "e2e", "failure", metrics.CategoryCodeTests},
		{"a credentials step", "Configure AWS credentials", "failure", metrics.CategoryConfigSecrets},
		{"a registry login step", "Docker login", "failure", metrics.CategoryConfigSecrets},
		{"an image pull step", "Pull base image", "failure", metrics.CategoryNetworkTimeouts},
		{"a dependency install step", "npm install", "failure", metrics.CategoryNetworkTimeouts},
		{"a runner setup step", "Set up job", "failure", metrics.CategoryInfrastructure},
		{"a cache restore step", "Restore cache", "failure", metrics.CategoryInfrastructure},
		{"matching ignores case", "RUN UNIT TESTS", "failure", metrics.CategoryCodeTests},
		{"credentials win over test in the name", "Test credentials", "failure", metrics.CategoryConfigSecrets},
		{"a keyword inside another word doesn't match", "Inspect image", "failure", metrics.CategoryUncategorised},
		{"latest isn't test", "Check latest version", "failure", metrics.CategoryUncategorised},
		{"author isn't auth", "Notify author", "failure", metrics.CategoryUncategorised},
		{"plurals match their stem", "Run integration tests", "failure", metrics.CategoryCodeTests},
		{"a hyphenated name splits into words", "golangci-lint", "failure", metrics.CategoryCodeTests},
		{"nothing matches", "Publish release", "failure", metrics.CategoryUncategorised},
		{"an empty name", "", "failure", metrics.CategoryUncategorised},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, metrics.CategorizeFailure(tt.step, tt.conclusion))
		})
	}
}

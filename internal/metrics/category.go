package metrics

import (
	"strings"
	"unicode"
)

// FailureCategory is a heuristic bucket for why a step failed, per
// pipeline-metrics/spec.md's "Failure grouping by step and category". It's
// inferred from the step's conclusion and name only -- no log is read -- so
// a failure no rule recognises is CategoryUncategorised, never a guess.
type FailureCategory string

// The failure categories a step can be assigned.
const (
	CategoryInfrastructure  FailureCategory = "infrastructure"
	CategoryCodeTests       FailureCategory = "code_tests"
	CategoryNetworkTimeouts FailureCategory = "network_timeouts"
	CategoryConfigSecrets   FailureCategory = "config_secrets"
	CategoryUncategorised   FailureCategory = "uncategorised"
)

// categoryRules are checked in order, so a name matching two rules takes the
// earlier one: "Test credentials" is a credentials problem before it's a test.
var categoryRules = []struct {
	category FailureCategory
	keywords []string
}{
	{CategoryConfigSecrets, []string{"credential", "secret", "token", "login", "authenticat", "authoriz", "kms", "oidc"}},
	{CategoryCodeTests, []string{"test", "spec", "lint", "vet", "typecheck", "e2e"}},
	{CategoryNetworkTimeouts, []string{"pull", "download", "fetch", "install", "registry"}},
	{CategoryInfrastructure, []string{"setup", "runner", "cache", "provision"}},
}

// CategorizeFailure assigns a failed step to a FailureCategory. A step that
// concluded timed_out is a timeout whatever it's called. Otherwise a keyword
// matches a word in the name by prefix, so "tests" finds "test" but "latest"
// and "inspect" don't.
func CategorizeFailure(stepName, conclusion string) FailureCategory {
	if conclusion == "timed_out" {
		return CategoryNetworkTimeouts
	}

	words := strings.FieldsFunc(strings.ReplaceAll(strings.ToLower(stepName), "set up", "setup"), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for _, rule := range categoryRules {
		for _, keyword := range rule.keywords {
			for _, word := range words {
				if strings.HasPrefix(word, keyword) {
					return rule.category
				}
			}
		}
	}

	return CategoryUncategorised
}

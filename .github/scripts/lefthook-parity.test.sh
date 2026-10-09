#!/usr/bin/env bash
# Checks that lefthook.yml keeps a hook for each CI check that has no
# exemption (rules/linting.md, "Everything CI runs should be runnable
# locally"), creates the Docker cache directories before the Go hooks mount
# them, and sets `output: [failure]`.
#
# Runs in the always-on `changes` job, so a check added to CI without a hook
# fails the pull request that adds it.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
file="${1:-$here/../../lefthook.yml}"
failures=0

expect() {
	if grep -Eq -- "$2" "$file"; then
		echo "ok:   $1"
	else
		echo "FAIL: $1 (lefthook.yml has no match for: $2)"
		failures=$((failures + 1))
	fi
}

expect "hadolint hook, guarded by the docker group" 'hook-guard\.sh --push docker -- .*hadolint'
expect "goreleaser check hook, guarded by the goreleaser group" 'hook-guard\.sh --push goreleaser --'
expect "goreleaser check runs" 'goreleaser/goreleaser:v[0-9.]+ check'
expect "redocly lint hook, guarded by the api group" 'hook-guard\.sh --push api -- bunx @redocly/cli lint'
expect "go.mod formatting check, guarded by the go group" 'go mod edit -fmt -print go\.mod'
expect "output: [failure] at the top level" '^output: \[failure\]$'
expect "govulncheck named CI-only" '^# CI-only.*govulncheck'
expect "bun audit named CI-only" '^# CI-only.*bun audit'

# Every hook that runs a Go image creates the cache directories first, or
# Docker creates them as root and the hook fails with permission denied.
images=$(grep -Ec 'golang:|golangci/golangci-lint:' "$file" || true)
makes=$(grep -c 'mkdir -p' "$file" || true)
if [ "$images" -gt 0 ] && [ "$makes" -eq "$images" ]; then
	echo "ok:   all $images Go image hooks create their cache directories"
else
	echo "FAIL: $images hooks run a Go image but $makes create cache directories"
	failures=$((failures + 1))
fi

# pre-commit judges only what the commit contains (rules/linting.md): every
# job there takes the staged files, none builds the whole tree, and none
# swallows its own failure. The section runs from `pre-commit:` to the next
# top-level key.
precommit=$(awk '/^pre-commit:/{f=1;next} /^[a-z]/{f=0} f' "$file")
if printf '%s\n' "$precommit" | grep -Eq '^    docker-build:'; then
	echo "FAIL: pre-commit has a docker-build job; builds belong in pre-push and CI"
	failures=$((failures + 1))
else
	echo "ok:   pre-commit has no whole-tree docker-build"
fi
if printf '%s\n' "$precommit" | grep -Eq '\|\| true'; then
	echo "FAIL: a pre-commit job ends in '|| true' and can never fail"
	failures=$((failures + 1))
else
	echo "ok:   no pre-commit job swallows its failure"
fi
expect "pre-push keeps the docker build" 'hook-guard\.sh --push docker -- sh -c .go build'

[ "$failures" -eq 0 ]

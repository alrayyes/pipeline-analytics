#!/usr/bin/env bash
# Checks that ci.yml uploads JUnit results to Codecov Test Analytics for both
# test suites, and gets the three details that fail silently when wrong
# (skills/repo-coverage, "Test Analytics"):
#
#  - `report_type` is underscore-separated; `report-type` only prints an
#    "Unexpected input" warning and the upload never becomes test results.
#  - the step runs after a failing test (`!cancelled()`), because a red run is
#    the case Test Analytics exists to explain.
#  - the Dependabot guard stays, since those runs get no CODECOV_TOKEN.
#
# Runs in the always-on `changes` job.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
file="${1:-$here/../workflows/ci.yml}"
failures=0

check() {
	if [ "$2" -eq "$3" ]; then
		echo "ok:   $1"
	else
		echo "FAIL: $1 (found $2, want $3)"
		failures=$((failures + 1))
	fi
}

check "two test_results uploads (go, frontend)" "$(grep -c 'report_type: test_results' "$file" || true)" 2
check "no hyphenated report-type" "$(grep -c 'report-type' "$file" || true)" 0
check "both uploads run after a failing test" "$(grep -c "if: \${{ !cancelled() && github.actor != 'dependabot\[bot\]' }}" "$file" || true)" 2
check "go suite writes JUnit" "$(grep -c 'gotestsum.*--junitfile' "$file" || true)" 1
check "frontend suite writes JUnit" "$(grep -c -- '--reporter=junit' "$here/../../web/package.json" || true)" 1

[ "$failures" -eq 0 ]

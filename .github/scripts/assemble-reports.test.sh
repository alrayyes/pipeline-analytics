#!/usr/bin/env bash
# Checks assemble-reports.sh lays the downloaded CI artifacts out as the
# published report tree (rules/published-reports.md): one JUnit file per
# runner, Cobertura coverage, Lighthouse output and an index, and that it
# refuses to assemble a partial tree, since a deploy replaces the whole site.
#
# Runs in the always-on `changes` job.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
script="$here/assemble-reports.sh"
failures=0

check() {
	if [ "$2" = "$3" ]; then
		echo "ok:   $1"
	else
		echo "FAIL: $1 (found '$2', want '$3')"
		failures=$((failures + 1))
	fi
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

artifacts="$tmp/artifacts"
mkdir -p "$artifacts/reports-go" "$artifacts/reports-frontend" "$artifacts/reports-e2e" "$artifacts/lighthouse-report"
echo '<testsuites name="go"/>' >"$artifacts/reports-go/junit.xml"
echo '<coverage line-rate="1"/>' >"$artifacts/reports-go/coverage.xml"
echo 'mode: atomic' >"$artifacts/reports-go/coverage.out"
echo '<html>go</html>' >"$artifacts/reports-go/coverage.html"
echo '<testsuites name="frontend"/>' >"$artifacts/reports-frontend/junit.xml"
echo '<coverage line-rate="1"/>' >"$artifacts/reports-frontend/coverage.xml"
echo 'TN:' >"$artifacts/reports-frontend/lcov.info"
echo '<testsuites name="e2e"/>' >"$artifacts/reports-e2e/e2e.xml"
echo '<html>lh</html>' >"$artifacts/lighthouse-report/lhr-1.html"
echo '{}' >"$artifacts/lighthouse-report/lhr-1.json"
echo '[]' >"$artifacts/lighthouse-report/manifest.json"

site="$tmp/site"
GITHUB_SHA=abc1234 "$script" "$artifacts" "$site" >/dev/null

exists() { [ -e "$site/reports/$1" ] && echo yes || echo no; }

check "go JUnit is tests/go.xml" "$(exists tests/go.xml)" yes
check "frontend JUnit is tests/frontend.xml" "$(exists tests/frontend.xml)" yes
check "e2e JUnit is tests/e2e.xml" "$(exists tests/e2e.xml)" yes
check "tests has an index" "$(exists tests/index.html)" yes
check "coverage.xml is the Go Cobertura file" "$(grep -c '<coverage ' "$site/reports/coverage/coverage.xml")" 1
check "Go's native coverage file ships beside it" "$(exists coverage/coverage.out)" yes
check "Go's HTML coverage view ships" "$(exists coverage/go.html)" yes
check "frontend coverage is under coverage/frontend" "$(exists coverage/frontend/coverage.xml)" yes
check "frontend's native lcov ships" "$(exists coverage/frontend/lcov.info)" yes
check "coverage has an index" "$(exists coverage/index.html)" yes
check "lighthouse html and json are published" "$(exists lighthouse/lhr-1.html)$(exists lighthouse/lhr-1.json)" yesyes
check "lighthouse has an index" "$(exists lighthouse/index.html)" yes
check "the landing page names the commit" "$(grep -c abc1234 "$site/reports/index.html")" 1

# A missing report must stop the deploy, not publish a site without it.
rm "$artifacts/reports-e2e/e2e.xml"

if "$script" "$artifacts" "$tmp/partial" >/dev/null 2>&1; then
	check "a missing report fails the assembly" ran failed
else
	check "a missing report fails the assembly" failed failed
fi

if [ "$failures" -gt 0 ]; then
	exit 1
fi

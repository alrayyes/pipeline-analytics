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
# The frontend upload spans web/junit.xml and web/coverage/*, so upload-artifact
# keeps the coverage/ folder; the other artifacts are flat.
mkdir -p "$artifacts/reports-go" "$artifacts/reports-frontend/coverage" "$artifacts/reports-e2e" "$artifacts/lighthouse-report"
echo '<testsuites name="go"/>' >"$artifacts/reports-go/junit.xml"
echo '<coverage line-rate="1"/>' >"$artifacts/reports-go/coverage.xml"
echo 'mode: atomic' >"$artifacts/reports-go/coverage.out"
echo '<html>go</html>' >"$artifacts/reports-go/coverage.html"
echo '<testsuites name="frontend"/>' >"$artifacts/reports-frontend/junit.xml"
echo '<coverage line-rate="1"/>' >"$artifacts/reports-frontend/coverage/coverage.xml"
echo 'TN:' >"$artifacts/reports-frontend/coverage/lcov.info"
echo '<testsuites name="e2e"/>' >"$artifacts/reports-e2e/e2e.xml"
# Lighthouse CI names every file for its URL and the minute it ran, so each
# deploy would rename them. The manifest it writes holds absolute runner paths.
lh="$artifacts/lighthouse-report"
write_lighthouse() {
	local stamp="$1"
	rm -f "$lh"/*
	echo '<html>login</html>' >"$lh/localhost-login-$stamp.report.html"
	echo '{"page":"login"}' >"$lh/localhost-login-$stamp.report.json"
	echo '<html>home</html>' >"$lh/localhost--$stamp.report.html"
	echo '{"page":"home"}' >"$lh/localhost--$stamp.report.json"
	echo '<html>repos</html>' >"$lh/localhost-repos-$stamp.report.html"
	echo '{"page":"repos"}' >"$lh/localhost-repos-$stamp.report.json"
	# A second, non-representative run of /repos, as numberOfRuns > 1 writes.
	echo '<html>repos-2</html>' >"$lh/localhost-repos-$stamp-b.report.html"
	echo '{"page":"repos-2"}' >"$lh/localhost-repos-$stamp-b.report.json"
	cat >"$lh/manifest.json" <<JSON
[
 {"url":"http://localhost:4191/login","isRepresentativeRun":true,"htmlPath":"/home/runner/work/x/web/.lighthouseci/localhost-login-$stamp.report.html","jsonPath":"/home/runner/work/x/web/.lighthouseci/localhost-login-$stamp.report.json","summary":{"performance":1}},
 {"url":"http://localhost:4191/","isRepresentativeRun":true,"htmlPath":"/home/runner/work/x/web/.lighthouseci/localhost--$stamp.report.html","jsonPath":"/home/runner/work/x/web/.lighthouseci/localhost--$stamp.report.json","summary":{"performance":1}},
 {"url":"http://localhost:4191/repos","isRepresentativeRun":false,"htmlPath":"/home/runner/work/x/web/.lighthouseci/localhost-repos-$stamp-b.report.html","jsonPath":"/home/runner/work/x/web/.lighthouseci/localhost-repos-$stamp-b.report.json","summary":{"performance":0.5}},
 {"url":"http://localhost:4191/repos","isRepresentativeRun":true,"htmlPath":"/home/runner/work/x/web/.lighthouseci/localhost-repos-$stamp.report.html","jsonPath":"/home/runner/work/x/web/.lighthouseci/localhost-repos-$stamp.report.json","summary":{"performance":1}}
]
JSON
}
write_lighthouse 2026_10_16_13_22_47

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
check "lighthouse html and json are published under the page's name" "$(exists lighthouse/login.report.html)$(exists lighthouse/login.report.json)" yesyes
check "the home page is called home" "$(exists lighthouse/home.report.html)$(exists lighthouse/home.report.json)" yesyes
check "only the representative run of a page is published" "$(cat "$site/reports/lighthouse/repos.report.json")" '{"page":"repos"}'
check "no report keeps its timestamped name" "$(find "$site/reports/lighthouse" -name '*[0-9][0-9][0-9][0-9]_[0-9][0-9]_*' | wc -l | tr -d ' ')" 0
lhm="$site/reports/lighthouse/manifest.json"
check "the manifest lists each page once, by its stable file names" "$(jq -r '[.[]|.htmlPath, .jsonPath]|sort|join(",")' "$lhm")" "home.report.html,home.report.json,login.report.html,login.report.json,repos.report.html,repos.report.json"
missing=0
for f in $(jq -r '.[]|.htmlPath, .jsonPath' "$lhm"); do
	[ -f "$site/reports/lighthouse/$f" ] || missing=$((missing + 1))
done
check "every file the manifest names exists" "$missing" 0
check "the manifest keeps each page's url and summary" "$(jq -r '.[]|select(.htmlPath=="repos.report.html")|"\(.url) \(.summary.performance)"' "$lhm")" "http://localhost:4191/repos 1"

# Another deploy, another timestamp: a manifest cached from the first still
# resolves, because the names did not move.
listing() { (cd "$1" && find . -type f | sort | tr '\n' ' '); }
before="$(listing "$site/reports/lighthouse")"
write_lighthouse 2026_10_17_09_00_01
"$script" "$artifacts" "$tmp/second" >/dev/null
check "the file names are the same after the next deploy" "$(listing "$tmp/second/reports/lighthouse")" "$before"
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

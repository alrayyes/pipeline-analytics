#!/usr/bin/env bash
# Lays the report artifacts the CI jobs uploaded out as <site>/reports/, the
# tree published at apis.ryankes.eu/pipeline-analytics/reports/
# (rules/published-reports.md). One job assembles it next to the docs, because
# a repo has one Pages deployment and a second replaces the first.
#
#   assemble-reports.sh <artifacts-dir> <site-dir>
#
# <artifacts-dir> holds one directory per downloaded artifact: reports-go,
# reports-frontend, reports-e2e and lighthouse-report. A missing file stops the
# assembly: a deploy replaces the whole site, so a partial tree would delete a
# report that was there. Its test is assemble-reports.test.sh.
set -euo pipefail

artifacts="${1:?usage: assemble-reports.sh <artifacts-dir> <site-dir>}"
site="${2:?usage: assemble-reports.sh <artifacts-dir> <site-dir>}"
out="$site/reports"

# take <from> <to>: copy one required file, or fail naming it.
take() {
	if [ ! -f "$artifacts/$1" ]; then
		echo "assemble-reports: missing $1" >&2
		exit 1
	fi

	mkdir -p "$(dirname "$out/$2")"
	cp "$artifacts/$1" "$out/$2"
}

# index <dir> <title>: write <dir>/index.html linking every other file in it.
index() {
	{
		echo "<!doctype html><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>$2</title>"
		echo "<h1>$2</h1><ul>"
		(cd "$1" && find . -type f ! -name index.html | sort | sed 's|^\./||') | while read -r f; do
			echo "<li><a href=\"$f\">$f</a></li>"
		done
		echo "</ul>"
	} >"$1/index.html"
}

rm -rf "$out"

take reports-go/junit.xml tests/go.xml
take reports-frontend/junit.xml tests/frontend.xml
take reports-e2e/e2e.xml tests/e2e.xml
take reports-go/coverage.xml coverage/coverage.xml
take reports-go/coverage.out coverage/coverage.out
take reports-go/coverage.html coverage/go.html
take reports-frontend/coverage/coverage.xml coverage/frontend/coverage.xml
take reports-frontend/coverage/lcov.info coverage/frontend/lcov.info

if [ ! -d "$artifacts/lighthouse-report" ]; then
	echo "assemble-reports: missing lighthouse-report" >&2
	exit 1
fi

# Lighthouse CI names each file for its URL and the minute it ran, so every
# deploy would rename them. GitHub Pages serves manifest.json and the index with
# max-age=600 behind a CDN, so for ten minutes a cached manifest would name files
# the deploy removed. Publish one pair per audited page under a name that does not
# change, <page>.report.html and <page>.report.json, and point the manifest at
# those. The page name is the URL's path, with the root called home.
lh_in="$artifacts/lighthouse-report"
lh_out="$out/lighthouse"

if [ ! -f "$lh_in/manifest.json" ]; then
	echo "assemble-reports: lighthouse-report has no manifest.json" >&2
	exit 1
fi

mkdir -p "$lh_out"

# One entry per URL: the representative run, or the first when none is marked.
pages="$(jq -c 'group_by(.url) | map((map(select(.isRepresentativeRun))[0] // .[0])) | .[]' "$lh_in/manifest.json")"

if [ -z "$pages" ]; then
	echo "assemble-reports: lighthouse-report has no HTML report" >&2
	exit 1
fi

manifest="[]"
while read -r entry; do
	url="$(jq -r .url <<<"$entry")"
	slug="$(sed -E 's|^[a-z]+://[^/]*||; s|[?#].*$||; s|^/+||; s|/+$||; s|/|-|g' <<<"$url")"
	[ -n "$slug" ] || slug="home"

	for kind in html json; do
		src="$lh_in/$(basename "$(jq -r ".${kind}Path" <<<"$entry")")"
		if [ ! -f "$src" ]; then
			echo "assemble-reports: missing $(basename "$src")" >&2
			exit 1
		fi
		cp "$src" "$lh_out/$slug.report.$kind"
	done

	manifest="$(jq --argjson e "$entry" --arg slug "$slug" \
		'. + [$e | .htmlPath = "\($slug).report.html" | .jsonPath = "\($slug).report.json"]' <<<"$manifest")"
done <<<"$pages"

jq . <<<"$manifest" >"$lh_out/manifest.json"

index "$out/tests" "Test results"
index "$out/coverage" "Coverage"
index "$out/lighthouse" "Lighthouse"

cat >"$out/index.html" <<HTML
<!doctype html><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>pipeline-analytics reports</title>
<h1>pipeline-analytics reports</h1>
<p>Commit ${GITHUB_SHA:-unknown}, built $(date -u +%Y-%m-%dT%H:%MZ).</p>
<ul>
<li><a href="tests/">Test results</a> (JUnit XML)</li>
<li><a href="coverage/">Coverage</a> (Cobertura XML)</li>
<li><a href="lighthouse/">Lighthouse</a></li>
</ul>
HTML

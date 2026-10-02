#!/usr/bin/env bash
# Tests changes.sh: which CI job groups a set of changed files turns on.
# Runs in the always-on `changes` job, so a wrong path fails CI before it
# silently skips a check that should have run.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
failures=0

# expect "<case name>" "<groups that must be true, space separated, or 'none'>" <files...>
# Every group not listed must be false.
expect() {
	local name="$1" want="$2"
	shift 2

	local got
	# grep exits 1 when nothing is true, which is a valid answer here.
	got="$(printf '%s\n' "$@" | "$here/changes.sh" | { grep '=true$' || true; } | sed 's/=true$//' | sort | tr '\n' ' ' | sed 's/ $//')"
	local wanted
	# shellcheck disable=SC2086 # $want is a space-separated list, split on purpose
	wanted="$(printf '%s\n' $want | { grep -v '^none$' || true; } | sort | tr '\n' ' ' | sed 's/ $//')"

	if [ "$got" != "$wanted" ]; then
		echo "FAIL: $name"
		echo "  want: ${wanted:-<none>}"
		echo "  got:  ${got:-<none>}"
		failures=$((failures + 1))
	else
		echo "ok:   $name"
	fi
}

ALL="api binary docker go goreleaser lint md mdlint ltex package vale web"

expect "a docs-only README edit" "md mdlint ltex vale" README.md
expect "a CONTRIBUTING edit" "md mdlint ltex vale" CONTRIBUTING.md
expect "a spec file edit is linted by ltex but not vale" "md mdlint ltex" openspec/specs/dashboard-ui/spec.md
expect "an openspec change is markdown but not prose-checked" "md mdlint" openspec/changes/x/proposal.md
expect "a Go source file" "binary docker go lint" internal/httpserver/server.go
expect "a Go test file" "binary docker go lint" internal/metrics/metrics_test.go
expect "go.mod" "binary docker go lint" go.mod
expect "the lint config" "lint md" .golangci.yml
expect "a frontend source file" "binary web" web/src/lib/format.ts
expect "a Svelte component" "binary web" web/src/routes/+page.svelte
expect "the frontend lockfile" "binary web" web/bun.lock
expect "the changelog, which the release page is built from" "binary md mdlint web" CHANGELOG.md
expect "the OpenAPI spec, which the MCP parity test reads, and Prettier checks as YAML" "api go md" openapi/openapi.yaml
expect "the API docs page" "api" docs/api/index.html
expect "the Dockerfile" "docker" Dockerfile
expect "the goreleaser config, also YAML" "goreleaser md" .goreleaser.yml
expect "root package.json" "api md mdlint package" package.json
expect "the web package.json" "binary package web" web/package.json
expect "a workflow yaml other than ci.yml" "md" .github/workflows/release.yml
expect "a vale style file" "md vale" styles/Google/Spelling.yml
expect "the ltex config" "ltex" .ltex.json
expect "an image under docs" "none" docs/screenshot-overview.png
expect "the workflow file itself runs everything" "$ALL" .github/workflows/ci.yml
expect "the filter script itself runs everything" "$ALL" .github/scripts/changes.sh
expect "its test runs everything" "$ALL" .github/scripts/changes.test.sh
expect "a mixed change turns on the union" "binary docker go lint md mdlint ltex vale" README.md internal/httpserver/server.go
expect "no changed files runs nothing" "none"

# An unknown base (a new branch, a force push) can't say what changed, so the
# caller asks for everything instead of filtering on a guess.
got="$("$here/changes.sh" --all | { grep -c '=true$' || true; })"
# shellcheck disable=SC2086 # $ALL is a space-separated list, split on purpose
if [ "$got" != "$(printf '%s\n' $ALL | wc -l)" ]; then
	echo "FAIL: --all turns on every group (got $got)"
	failures=$((failures + 1))
else
	echo "ok:   --all turns on every group"
fi

[ "$failures" -eq 0 ] || { echo "$failures case(s) failed"; exit 1; }
echo "all cases passed"

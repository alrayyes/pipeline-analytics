#!/usr/bin/env bash
# Tests hook-guard.sh: a lefthook job runs its command only when changes.sh
# says the job's CI group changed. HOOK_FILES stands in for the git diff.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
failures=0

# expect "<case>" <ran|skipped> "<group>" <files...>
expect() {
	local name="$1" want="$2" group="$3"
	shift 3

	local got=skipped
	if [ "$(printf '%s\n' "$@" | HOOK_FILES=- "$here/hook-guard.sh" "$group" -- echo ran)" = ran ]; then
		got=ran
	fi

	if [ "$got" != "$want" ]; then
		echo "FAIL: $name (want $want, got $got)"
		failures=$((failures + 1))
	else
		echo "ok:   $name"
	fi
}

expect "go tests run for a spec-only change" ran go openapi/openapi.yaml
expect "go tests run for go.sum" ran go go.sum
expect "lint runs for the lint config" ran lint .golangci.yml
expect "web checks run for the changelog" ran web CHANGELOG.md
expect "go tests skip for a docs-only change" skipped go README.md
expect "web checks skip for a docs-only change" skipped web README.md
expect "markdown lint runs for a docs-only change" ran mdlint README.md
expect "an unknown group fails loudly" skipped nonesuch README.md

if "$here/hook-guard.sh" nonesuch -- true </dev/null 2>/dev/null; then
	echo "FAIL: unknown group should exit non-zero"
	failures=$((failures + 1))
fi

[ "$failures" -eq 0 ]

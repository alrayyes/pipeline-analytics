#!/usr/bin/env bash
# Checks that every workflow which installs or runs bun tooling sets
# LEFTHOOK=0 (rules/tooling.md, "Every pipeline sets LEFTHOOK=0"): the root
# `prepare` script is `lefthook install`, so a pipeline install would
# otherwise wire the repo's git hooks into a runner that may not carry the
# tools they expect, and a job that commits would fire them.
#
# Runs in the always-on `changes` job, so a new workflow that forgets it fails
# the pull request that adds it. It has to be the workflow-level `env:` (two
# spaces in), so no job or step is left to remember it separately.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dir="${1:-$here/../workflows}"
failures=0

for file in "$dir"/*.yml "$dir"/*.yaml; do
	[ -e "$file" ] || continue

	if grep -Eq 'bun install|bunx |bun run' "$file"; then
		if grep -Eq '^  LEFTHOOK:[[:space:]]*"?0"?[[:space:]]*$' "$file"; then
			echo "ok:   $(basename "$file") sets LEFTHOOK=0"
		else
			echo "FAIL: $(basename "$file") runs bun but doesn't set LEFTHOOK: \"0\" in its top-level env"
			failures=$((failures + 1))
		fi
	fi
done

[ "$failures" -eq 0 ]

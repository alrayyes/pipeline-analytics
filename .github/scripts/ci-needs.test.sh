#!/usr/bin/env bash
# Checks that ci.yml's deploy job waits for every other job in the workflow
# (rules/ci.md, "A deploy job depends on every other job in its own
# pipeline"): a job added later doesn't join a `needs:` list by itself, and a
# red job that finishes after the deploy already shipped is a shipped bug.
# Runs in the always-on `changes` job, so adding a job without listing it fails
# the pull request that adds it.
#
# It reads the workflow as text: job ids are the two-space-indented keys under
# `jobs:`, and the deploy's `needs:` has to be a flow list (`[a, b]`), which
# Prettier may wrap over several lines.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
workflow="${1:-$here/../workflows/ci.yml}"
deploy="${2:-pages-deploy}"

jobs="$(awk '/^jobs:/{f=1;next} f&&/^  [a-z0-9-]+:$/{gsub(/[: ]/,"");print}' "$workflow")"

needs="$(awk -v job="$deploy" '
	$0 == "  " job ":" {f=1; next}
	f && /^  [a-z0-9-]+:$/ {exit}
	f && /^    needs:/ {c=1; sub(/^    needs:[ ]*/, "")}
	c {buf = buf " " $0; if ($0 ~ /\]/) {gsub(/[\[\],]/, " ", buf); print buf; exit}}
' "$workflow")"

if [ -z "$needs" ]; then
	echo "FAIL: $deploy has no flow-list 'needs: [...]' in $workflow"
	exit 1
fi

missing=""
for job in $jobs; do
	[ "$job" = "$deploy" ] && continue

	case " $needs " in
	*" $job "*) ;;
	*) missing="$missing $job" ;;
	esac
done

if [ -n "$missing" ]; then
	echo "FAIL: $deploy doesn't wait for:$missing"
	echo "  add them to its needs: list in $workflow"
	exit 1
fi

echo "ok:   $deploy waits for every other job"

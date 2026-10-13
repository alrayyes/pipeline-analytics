#!/usr/bin/env bash
# Tests wait-for-artifact.sh: it returns once the run's artifact list names the
# artifact, and fails if the list never does. A stub lister stands in for the
# API; it names the artifact from its Nth call on.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
failures=0

# Writes a lister that prints "other" until call <n>, then also "github-pages".
stub() {
	cat >"$tmp/list.sh" <<STUB
#!/usr/bin/env bash
count=\$(cat "$tmp/count" 2>/dev/null || echo 0)
count=\$((count + 1))
echo "\$count" >"$tmp/count"
echo other
if [ "\$count" -ge $1 ]; then echo github-pages; fi
STUB
	chmod +x "$tmp/list.sh"
	rm -f "$tmp/count"
}

# expect "<case>" <pass|fail> <first-call-that-lists-it> <tries>
expect() {
	local name="$1" want="$2" first="$3" tries="$4" got=fail
	stub "$first"
	if WAIT_TRIES="$tries" WAIT_SLEEP=0 LIST_ARTIFACTS="$tmp/list.sh" \
		"$here/wait-for-artifact.sh" github-pages >/dev/null 2>&1; then
		got=pass
	fi

	if [ "$got" != "$want" ]; then
		echo "FAIL: $name (want $want, got $got)"
		failures=$((failures + 1))
	else
		echo "ok:   $name"
	fi
}

expect "an artifact listed at once passes" pass 1 5
expect "an artifact listed on the third look passes" pass 3 5
expect "an artifact listed on the last look passes" pass 5 5
expect "an artifact that never lists fails" fail 99 5

if "$here/wait-for-artifact.sh" </dev/null 2>/dev/null; then
	echo "FAIL: a missing name should exit non-zero"
	failures=$((failures + 1))
fi

exit "$failures"

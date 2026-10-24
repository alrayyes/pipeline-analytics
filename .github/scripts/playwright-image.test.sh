#!/usr/bin/env bash
# Checks that the jobs which run a browser use the prebuilt CI image
# (ghcr.io/alrayyes/ci-playwright) and that its tag is the @playwright/test
# version web/bun.lock locks. The image's browsers only work with the library
# version they were built for, and a container image can't read the lockfile,
# so the tag is written out in each workflow; this fails when one drifts, is
# missing its digest, or when a workflow goes back to `playwright install`.
#
# It first proves the check itself against fixtures (a matching pair passes,
# each kind of drift fails), then runs it on the real files. Runs in the
# always-on `changes` job and in the pre-push hooks.
#
#   playwright-image.test.sh [workflows-dir] [lockfile]
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
failures=0

# check <workflows-dir> <lockfile>: prints what is wrong, returns non-zero.
check() {
	local dir="$1" lockfile="$2" bad=0 lib file count image
	local images=()

	lib="$(grep -Eo '"@playwright/test@[0-9]+\.[0-9]+\.[0-9]+"' "$lockfile" | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+' | head -n1 || true)"
	if [ -z "$lib" ]; then
		echo "  $lockfile locks no exact @playwright/test version"
		return 1
	fi

	for file in "$dir"/*.yml "$dir"/*.yaml; do
		[ -e "$file" ] || continue
		if grep -Eq 'playwright install' "$file"; then
			echo "  $(basename "$file") still runs 'playwright install'; the image has the browsers"
			bad=1
		fi
		while IFS= read -r image; do
			images+=("$image")
			if ! [[ $image =~ ^ghcr\.io/alrayyes/ci-playwright:([0-9]+\.[0-9]+\.[0-9]+)@sha256:[0-9a-f]{64}$ ]]; then
				echo "  $(basename "$file"): '$image' isn't pinned by an exact version and a digest (X.Y.Z@sha256:...)"
				bad=1
			elif [ "${BASH_REMATCH[1]}" != "$lib" ]; then
				echo "  $(basename "$file"): image is for Playwright ${BASH_REMATCH[1]} but $lockfile locks $lib"
				bad=1
			fi
			if ! grep -Eq '^[[:space:]]+options:.*--ipc=host' "$file"; then
				echo "  $(basename "$file"): the container needs 'options: --ipc=host' for Chromium"
				bad=1
			fi
		done < <(grep -E '^[[:space:]]+image: ghcr\.io/alrayyes/ci-playwright' "$file" | sed -E 's/^[[:space:]]+image:[[:space:]]+//' || true)
	done

	count="${#images[@]}"
	if [ "$count" -lt 1 ]; then
		echo "  no job in $dir uses the ci-playwright image"
		bad=1
	elif [ "$(printf '%s\n' "${images[@]}" | sort -u | wc -l)" -gt 1 ]; then
		echo "  the jobs use different images: ${images[*]}"
		bad=1
	fi
	return "$bad"
}

# expect <name> pass|fail <workflows-dir> <lockfile>
expect() {
	local name="$1" want="$2" out status=0
	out="$(check "$3" "$4")" || status=$?
	if { [ "$want" = pass ] && [ "$status" -eq 0 ]; } || { [ "$want" = fail ] && [ "$status" -ne 0 ]; }; then
		echo "ok:   $name"
	else
		echo "FAIL: $name (wanted $want)"
		[ -n "$out" ] && echo "$out"
		failures=$((failures + 1))
	fi
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
digest="sha256:$(printf 'a%.0s' {1..64})"

fixture() { # fixture <name> <image> [extra step line]
	mkdir -p "$tmp/$1"
	printf 'jobs:\n  e2e:\n    container:\n      image: %s\n      options: %s\n    steps:\n      - run: echo hi\n%s\n' "$2" "${IPC:---ipc=host}" "${3:-}" >"$tmp/$1/ci.yml"
}
printf '    "@playwright/test": ["@playwright/test@1.64.1", "", {}],\n' >"$tmp/lock"

fixture ok "ghcr.io/alrayyes/ci-playwright:1.64.1@$digest"
fixture older "ghcr.io/alrayyes/ci-playwright:1.63.0@$digest"
fixture nodigest "ghcr.io/alrayyes/ci-playwright:1.64.1"
fixture mcr "mcr.microsoft.com/playwright:v1.64.1-noble@$digest"
fixture reinstall "ghcr.io/alrayyes/ci-playwright:1.64.1@$digest" '      - run: bunx playwright install --with-deps chromium'
IPC=--init fixture noipc "ghcr.io/alrayyes/ci-playwright:1.64.1@$digest"
mkdir -p "$tmp/none" && printf 'jobs: {}\n' >"$tmp/none/ci.yml"
fixture split "ghcr.io/alrayyes/ci-playwright:1.64.1@$digest"
printf 'jobs:\n  s:\n    container:\n      image: ghcr.io/alrayyes/ci-playwright:1.64.1@sha256:%s\n' "$(printf 'b%.0s' {1..64})" >"$tmp/split/release.yml"

expect "a matching tag and digest passes" pass "$tmp/ok" "$tmp/lock"
expect "a tag older than the lockfile fails" fail "$tmp/older" "$tmp/lock"
expect "a tag with no digest fails" fail "$tmp/nodigest" "$tmp/lock"
expect "another registry's image fails" fail "$tmp/mcr" "$tmp/lock"
expect "a leftover playwright install fails" fail "$tmp/reinstall" "$tmp/lock"
expect "a container without --ipc=host fails" fail "$tmp/noipc" "$tmp/lock"
expect "no job on the image fails" fail "$tmp/none" "$tmp/lock"
expect "two workflows on different digests fail" fail "$tmp/split" "$tmp/lock"

expect "the real workflows match web/bun.lock" pass "${1:-$here/../workflows}" "${2:-$here/../../web/bun.lock}"

[ "$failures" -eq 0 ]

#!/usr/bin/env bash
# Tests scripts/lint-mechanics.sh: run bare (CI) it hands LTeX the whole
# prose list; given the files in a push (pre-push) it hands over only the
# Markdown among them that CI would check, and none at all when there is none.
# A fake `docker` records what reaches the ltex-cli-plus image, so no JVM
# starts. The hook's glob is checked statically.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
script="${SCRIPT:-$root/scripts/lint-mechanics.sh}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
failures=0

fail() {
	echo "FAIL: $1"
	failures=$((failures + 1))
}

if [ ! -x "$script" ]; then
	echo "FAIL: $script is missing or not executable"
	exit 1
fi

# The fake records the image it was given and the arguments after the
# --client-configuration flag, which are the files LTeX would check.
mkdir -p "$tmp/bin" "$tmp/repo/scripts"
cat >"$tmp/bin/docker" <<FAKE
#!/usr/bin/env bash
printf '%s\n' "\$@" | grep -F 'ltex-cli-plus@sha256:' >"$tmp/image" || true
printf '%s\n' "\$@" | sed -n '/^--client-configuration=/,\$p' | tail -n +2 >"$tmp/seen"
FAKE
chmod +x "$tmp/bin/docker"
cp "$script" "$tmp/repo/scripts/lint-mechanics.sh"

cd "$tmp/repo"
git init -q
echo '{}' >.ltex.json
mkdir -p .github openspec/specs/a openspec/changes/x docs
for f in README.md CONTRIBUTING.md SECURITY.md .github/PULL_REQUEST_TEMPLATE.md \
	openspec/specs/a/spec.md CHANGELOG.md docs/guide.md openspec/changes/x/proposal.md; do
	echo hi >"$f"
done
echo hi >README.md.bak
git add -A

# run [args...]: what the fake LTeX was handed, sorted and space-joined, or
# "none" when it never ran. The exit status lands in $tmp/status.
run() {
	rm -f "$tmp/seen" "$tmp/image"
	set +e
	PATH="$tmp/bin:$PATH" scripts/lint-mechanics.sh "$@" >"$tmp/out" 2>&1
	echo "$?" >"$tmp/status"
	set -e
	if [ -f "$tmp/seen" ]; then LC_ALL=C sort "$tmp/seen" | tr '\n' ' '; else echo none; fi
}

all=".github/PULL_REQUEST_TEMPLATE.md CONTRIBUTING.md README.md SECURITY.md openspec/specs "

[ "$(run)" = "$all" ] || fail "a bare run is not the whole prose list: $(run)"
[ -s "$tmp/image" ] || fail "the pinned ltex-cli-plus image was not used"
[ "$(run README.md)" = "README.md " ] || fail "a pushed file is not the only one checked: $(run README.md)"
[ "$(run README.md src/main.go)" = "README.md " ] || fail "a non-Markdown file reached ltex"
[ "$(run README.md CONTRIBUTING.md)" = "CONTRIBUTING.md README.md " ] || fail "two pushed files were not both checked"
[ "$(run openspec/specs/a/spec.md)" = "openspec/specs/a/spec.md " ] || fail "a pushed spec was not checked by itself"
[ "$(run CHANGELOG.md docs/guide.md openspec/changes/x/proposal.md)" = "none" ] || fail "Markdown outside CI's list was checked"
[ "$(run README.md.bak)" = "none" ] || fail "a file that merely contains a Markdown name was checked"
[ "$(run SECURITY.md deleted.md)" = "SECURITY.md " ] || fail "a deleted file reached ltex"
run docs/guide.md >/dev/null
[ "$(cat "$tmp/status")" = 0 ] || fail "a push with no prose must exit 0"
grep -Fq 'No prose in this push for ltex to check.' "$tmp/out" || fail "a push with no prose did not say so"
[ "$(run README.md .ltex.json)" = "$all" ] || fail "Markdown plus .ltex.json did not check everything"
[ "$(run .ltex.json)" = "$all" ] || fail "a config-only push did not check everything"
[ "$(run scripts/lint-mechanics.sh)" = "$all" ] || fail "a script-only push did not check everything"

cd "$root"
job=$(awk '/^    prose-mechanics:$/{p=1;print;next} /^    [a-z-]+:$/{p=0} p' lefthook.yml)
grep -q 'glob:' <<<"$job" || fail "the ltex hook has no glob, so every push starts a JVM"
grep -Fq '{push_files}' <<<"$job" || fail "the ltex hook does not pass the pushed files"
grep -Fq '.ltex.json' <<<"$job" || fail "the ltex hook ignores a change to .ltex.json"
grep -Fq 'scripts/lint-mechanics.sh' <<<"$job" || fail "the ltex hook does not run lint-mechanics.sh"
grep -Fq 'scripts/lint-mechanics.sh' .github/workflows/ci.yml || fail "CI does not run lint-mechanics.sh"

[ "$failures" -eq 0 ] && echo "ok:   lint-mechanics"
[ "$failures" -eq 0 ]

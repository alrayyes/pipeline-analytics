#!/usr/bin/env bash
# Runs a command only when changes.sh says the given CI group changed, so a
# lefthook job and its CI job can't disagree about what a change covers.
#
#   hook-guard.sh [--push] <group> -- <command...>
#
# Staged files decide by default (pre-commit); --push compares HEAD with
# origin/main (pre-push), and runs the command when there's no base to
# compare with. HOOK_FILES=- reads the file list from stdin, for the test.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mode=staged
if [ "${1:-}" = "--push" ]; then
	mode=push
	shift
fi
group="${1:?usage: hook-guard.sh [--push] <group> -- <command...>}"
shift
[ "${1:-}" = "--" ] && shift

if [ "${HOOK_FILES:-}" = "-" ]; then
	groups="$("$here/changes.sh")"
elif [ "$mode" = push ] && ! git rev-parse --verify -q origin/main >/dev/null; then
	groups="$("$here/changes.sh" --all </dev/null)"
elif [ "$mode" = push ]; then
	groups="$(git diff --name-only origin/main...HEAD | "$here/changes.sh")"
else
	groups="$(git diff --cached --name-only --diff-filter=ACMR | "$here/changes.sh")"
fi

case "$(printf '%s\n' "$groups" | grep -E "^${group}=" || true)" in
"${group}=true") exec "$@" ;;
"${group}=false") exit 0 ;;
*)
	echo "hook-guard: '$group' isn't a group in changes.sh" >&2
	exit 2
	;;
esac

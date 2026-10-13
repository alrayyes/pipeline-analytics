#!/usr/bin/env bash
# Waits until this run's artifact list names <name>. actions/deploy-pages looks
# the Pages artifact up through that list, which can lag the upload by a moment:
# it looked 0.3s after the upload finished and found nothing (#606). Its test
# is wait-for-artifact.test.sh.
#
#   wait-for-artifact.sh <name>
#
# WAIT_TRIES (default 12) and WAIT_SLEEP (default 5, seconds) bound the wait.
# LIST_ARTIFACTS names a command that prints one artifact name per line; the
# default asks the API about this run, which needs `actions: read`.
set -euo pipefail

name="${1:?usage: wait-for-artifact.sh <name>}"
tries="${WAIT_TRIES:-12}"
pause="${WAIT_SLEEP:-5}"

list() {
	if [ -n "${LIST_ARTIFACTS:-}" ]; then
		"$LIST_ARTIFACTS"
	else
		gh api --paginate "repos/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID/artifacts" \
			--jq '.artifacts[].name'
	fi
}

for ((attempt = 1; attempt <= tries; attempt++)); do
	if list | grep -qx -- "$name"; then
		echo "wait-for-artifact: $name is listed (look $attempt of $tries)"
		exit 0
	fi

	sleep "$pause"
done

echo "wait-for-artifact: $name is not in the run's artifact list after $tries looks" >&2
exit 1

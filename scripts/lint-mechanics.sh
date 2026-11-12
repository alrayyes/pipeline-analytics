#!/usr/bin/env bash
# Runs LTeX+ (grammar and spelling) over the prose CI gates, the one entry
# point the pre-push hook and CI both call (markdown.md, "Running the two
# together"). Run bare, as CI does, it checks every document on the list.
# The hook passes the files in the push: only the Markdown among them that
# the list covers is checked, so a push with no prose never starts a JVM. A
# push that includes .ltex.json or this script, either of which can change
# the verdict on any file, checks everything, as CI would.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# renovate: datasource=docker depName=ghcr.io/alrayyes/ltex-cli-plus
IMAGE="ghcr.io/alrayyes/ltex-cli-plus@sha256:00c30809802234d6c43b0ce0b295075bd0aff9c576aaba57433bc8baf314398c"

all=(README.md CONTRIBUTING.md SECURITY.md .github/PULL_REQUEST_TEMPLATE.md openspec/specs)
files=("${all[@]}")

if [ "$#" -gt 0 ]; then
	if printf '%s\n' "$@" | grep -Eq '^(\.ltex\.json|scripts/lint-mechanics\.sh)$'; then
		: # keep the whole list
	else
		# Same set as the list above, as tracked files: a deleted path
		# isn't tracked, so it drops out here.
		covered=$(git ls-files -- README.md CONTRIBUTING.md SECURITY.md \
			.github/PULL_REQUEST_TEMPLATE.md 'openspec/specs/*.md')
		mapfile -t files < <(printf '%s\n' "$@" | grep -Fx -f <(printf '%s\n' "$covered") || true)
		if [ "${#files[@]}" -eq 0 ]; then
			echo "No prose in this push for ltex to check."
			exit 0
		fi
	fi
fi

echo "Checking: ${files[*]}"
exec docker run --rm -v "$PWD:/data" -w /data "$IMAGE" \
	--client-configuration=/data/.ltex.json "${files[@]}"

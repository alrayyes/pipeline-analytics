#!/usr/bin/env bash
# Reads changed file paths on stdin and prints one `<group>=true|false` line
# per CI job group, for ci.yml's `changes` job to hand to the jobs that
# follow. `--all` turns every group on, for when the base is unknown (a new
# branch, a force push) and filtering would be a guess.
#
# A group lists the files its job covers, the tool's own config and
# lockfile, and the workflow and this script, so editing the pipeline runs
# everything. When adding a path, add a case to changes.test.sh first.
set -euo pipefail

GROUPS_LIST=(api binary docker go goreleaser lint md mdlint ltex package vale web)

# Go sources plus what a Go build or test reads: the embedded frontend and
# the OpenAPI spec (the MCP parity test parses it).
GO_SOURCES='\.go$|^go\.(mod|sum)$'

declare -A pattern=(
	[go]="$GO_SOURCES|^openapi/|^internal/webassets/|^codecov\.yml$"
	[lint]="$GO_SOURCES|^\.golangci\.ya?ml$"
	[web]='^web/|^CHANGELOG\.md$'
	# e2e and lighthouse build the real binary, so they cover the frontend and Go.
	[binary]="$GO_SOURCES|^web/|^CHANGELOG\.md$|^internal/webassets/"
	[docker]="$GO_SOURCES|^Dockerfile$|^\.dockerignore$|^\.hadolint"
	[api]='^openapi/|^docs/api/|^redocly\.ya?ml$|^\.spectral\.ya?ml$|^package\.json$|^bun\.lock$'
	[package]='^package\.json$|^web/package\.json$|^bun\.lock$'
	[goreleaser]='^\.goreleaser\.ya?ml$'
	[md]='\.(md|ya?ml)$|^\.prettier|^package\.json$|^bun\.lock$'
	[mdlint]='\.md$|^\.markdownlint|^package\.json$|^bun\.lock$'
	[ltex]='^(README|CONTRIBUTING|SECURITY)\.md$|^\.github/PULL_REQUEST_TEMPLATE\.md$|^openspec/specs/|^\.ltex\.json$'
	[vale]='^(README|CONTRIBUTING|SECURITY)\.md$|^\.github/PULL_REQUEST_TEMPLATE\.md$|^styles/|^\.vale\.ini$|^scripts/lint-prose\.sh$'
)

# The pipeline itself: editing it must run every job, or a broken filter
# would hide its own mistake.
PIPELINE='^\.github/workflows/ci\.yml$|^\.github/scripts/changes(\.test)?\.sh$'

files=""
run_all=false

if [ "${1:-}" = "--all" ]; then
	run_all=true
else
	files="$(cat)"
	if printf '%s\n' "$files" | grep -Eq "$PIPELINE"; then
		run_all=true
	fi
fi

for group in "${GROUPS_LIST[@]}"; do
	if $run_all || printf '%s\n' "$files" | grep -Eq "${pattern[$group]}"; then
		echo "$group=true"
	else
		echo "$group=false"
	fi
done

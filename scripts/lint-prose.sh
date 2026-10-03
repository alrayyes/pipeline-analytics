#!/usr/bin/env bash
# Runs Vale over the prose it gates, the one entry point the hooks and CI
# both call (markdown.md, "Running Vale"). With file arguments it checks
# those; with none, the same list the CI job checks.
#
# Uses a `vale` on PATH (CI's job runs inside the image, so it finds one).
# Without one, and with Docker, it runs the same pinned image, so a
# contributor's machine needs nothing installed.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# renovate: datasource=docker depName=jdkato/vale
VALE_IMAGE="jdkato/vale:v3.21.0@sha256:8a5a05b5c1751c6a6132ca459d67ffe3ad2cc338abc5795ee634470dd39de9c4"

if [ "$#" -eq 0 ]; then
	set -- README.md CONTRIBUTING.md SECURITY.md .github/PULL_REQUEST_TEMPLATE.md
fi

if command -v vale >/dev/null 2>&1; then
	vale sync
	exec vale "$@"
fi

if ! command -v docker >/dev/null 2>&1; then
	echo "lint-prose: neither vale nor docker is installed" >&2
	exit 127
fi

# --entrypoint "": the image's entrypoint is vale itself, and this runs a
# shell so the sync and the check share one container.
exec docker run --rm --user "$(id -u):$(id -g)" \
	-v "$PWD:/work" -w /work --entrypoint "" "$VALE_IMAGE" \
	sh -c 'vale sync && vale "$@"' _ "$@"

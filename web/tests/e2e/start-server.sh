#!/usr/bin/env bash
# Builds the frontend and the Go binary (which go:embeds it), then runs the
# real server for Playwright to test against -- same process shape as
# production, no dev-server/API proxy to diverge from it.
set -euo pipefail

cd "$(dirname "$0")/../../.."

(cd web && bun run build)

binary="$(mktemp -u /tmp/pipeline-analytics-e2e-XXXXXX)"
go build -o "$binary" ./cmd/pipeline-analytics

db="$(mktemp -u /tmp/pipeline-analytics-e2e-XXXXXX.db)"
key="$(openssl rand -hex 32)"

exec "$binary" serve \
	--addr ":4173" \
	--db "$db" \
	--callback-url "http://localhost:4173" \
	--encryption-key "$key" \
	--reconcile-interval "1h"

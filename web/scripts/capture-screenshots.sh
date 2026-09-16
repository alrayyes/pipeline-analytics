#!/usr/bin/env bash
# Builds the frontend and the Go binary (which go:embeds it), then runs the
# real server for capture-screenshots.mjs to screenshot -- same process
# shape as production, same reasoning as tests/e2e/start-server.sh.
set -euo pipefail

cd "$(dirname "$0")/../.."

(cd web && bun run build)

binary="$(mktemp -u /tmp/pipeline-analytics-screenshots-XXXXXX)"
go build -o "$binary" ./cmd/pipeline-analytics

db="$(mktemp -u /tmp/pipeline-analytics-screenshots-XXXXXX.db)"
key="$(openssl rand -hex 32)"

"$binary" serve \
	--addr ":4190" \
	--db "$db" \
	--callback-url "http://localhost:4190" \
	--encryption-key "$key" \
	--reconcile-interval "1h" &
server_pid=$!
trap 'kill "$server_pid" 2>/dev/null; rm -f "$db"' EXIT

for _ in $(seq 1 20); do
	curl -sf http://localhost:4190/healthz >/dev/null && break
	sleep 0.5
done

(cd web && bun scripts/capture-screenshots.mjs)

#!/usr/bin/env bash
# Builds the frontend and the Go binary (which go:embeds it), then runs the
# real server for lhci to audit -- same process shape as production, same
# reasoning as web/scripts/capture-screenshots.sh and tests/e2e's
# global-setup.ts/isolated-server.ts.
#
# /login is audited signed out. The pages behind the passkey
# (docs/adr/0002-webauthn-only-authentication.md) are audited signed in:
# scripts/lighthouse-auth.cjs registers a passkey through a virtual
# authenticator first, and check-lighthouse-reports.ts fails the run if any
# page was bounced to /login instead.
set -euo pipefail

cd "$(dirname "$0")/../.."

(cd web && bun run build)

binary="$(mktemp -u /tmp/pipeline-analytics-lighthouse-XXXXXX)"
go build -o "$binary" ./cmd/pipeline-analytics

db="$(mktemp -u /tmp/pipeline-analytics-lighthouse-XXXXXX.db)"
key="$(openssl rand -hex 32)"

"$binary" serve \
	--addr ":4191" \
	--db "$db" \
	--callback-url "http://localhost:4191" \
	--encryption-key "$key" \
	--reconcile-interval "1h" &
server_pid=$!
trap 'kill "$server_pid" 2>/dev/null; rm -f "$db"' EXIT

for _ in $(seq 1 20); do
	curl -sf http://localhost:4191/healthz >/dev/null && break
	sleep 0.5
done

# lhci finds Chrome through puppeteer, and puppeteer-core (all the
# puppeteerScript needs) can't say where it is, so name it. CHROME_PATH
# already set wins.
if [ -z "${CHROME_PATH:-}" ]; then
	CHROME_PATH="$(cd web && node -e "console.log(require('chrome-launcher').Launcher.getFirstInstallation() ?? '')")"
	export CHROME_PATH
fi

(cd web && bunx lhci autorun && bun scripts/check-lighthouse-reports.ts)

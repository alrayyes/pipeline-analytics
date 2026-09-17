import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

// Every test starts its own server from this one binary (fixtures.ts /
// isolated-server.ts) instead of the whole run sharing a single server
// process -- building it once here, rather than per test, is what keeps
// that from paying a full `vite build` + `go build` on every test and
// every retry.
export const BINARY_PATH = '/tmp/pipeline-analytics-e2e-binary';

const repoRoot = path.resolve(
	path.dirname(fileURLToPath(import.meta.url)),
	'../../..',
);

export default function globalSetup(): void {
	execFileSync('bun', ['run', 'build'], {
		cwd: path.join(repoRoot, 'web'),
		stdio: 'inherit',
	});

	execFileSync('go', ['build', '-o', BINARY_PATH, './cmd/pipeline-analytics'], {
		cwd: repoRoot,
		stdio: 'inherit',
	});
}

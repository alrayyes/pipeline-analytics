import { type ChildProcess, spawn } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { mkdtempSync, rmSync } from 'node:fs';
import net from 'node:net';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { BINARY_PATH } from './global-setup.js';

export interface IsolatedServer {
	baseURL: string;
	stop(): void;
}

// Lets the OS assign a free ephemeral port, then releases it immediately --
// the small window before the real server binds it is the same race every
// "find a free port" test helper accepts.
async function getFreePort(): Promise<number> {
	return new Promise((resolve, reject) => {
		const srv = net.createServer();
		srv.listen(0, () => {
			const address = srv.address();

			if (address === null || typeof address === 'string') {
				reject(new Error('could not determine a free port'));

				return;
			}

			const { port } = address;
			srv.close(() => resolve(port));
		});
		srv.on('error', reject);
	});
}

async function waitUntilHealthy(
	baseURL: string,
	proc: ChildProcess,
): Promise<void> {
	const deadline = Date.now() + 15_000;

	while (Date.now() < deadline) {
		if (proc.exitCode !== null) {
			throw new Error(`server exited early with code ${proc.exitCode}`);
		}

		try {
			const res = await fetch(`${baseURL}/healthz`);
			if (res.ok) return;
		} catch {
			// Not listening yet.
		}

		await new Promise((resolve) => setTimeout(resolve, 100));
	}

	throw new Error(`server at ${baseURL} did not become healthy in time`);
}

// Starts a fresh pipeline-analytics server against its own temp SQLite DB
// and port. Every caller gets a clean slate -- including a Playwright
// retry of a failed test (#168), which a single server shared for the
// whole run can't give it: a retry used to inherit whatever state (a
// registered passkey, tracked repos) the failed attempt left behind.
export async function startIsolatedServer(): Promise<IsolatedServer> {
	const dir = mkdtempSync(path.join(tmpdir(), 'pipeline-analytics-e2e-'));
	const dbPath = path.join(dir, 'pipeline-analytics.db');
	const port = await getFreePort();
	const baseURL = `http://localhost:${port}`;
	const key = randomBytes(32).toString('hex');

	const proc = spawn(
		BINARY_PATH,
		[
			'serve',
			'--addr',
			`:${port}`,
			'--db',
			dbPath,
			'--callback-url',
			baseURL,
			'--encryption-key',
			key,
			'--reconcile-interval',
			'1h',
		],
		{ stdio: 'ignore' },
	);

	await waitUntilHealthy(baseURL, proc);

	return {
		baseURL,
		stop() {
			proc.kill();
			rmSync(dir, { recursive: true, force: true });
		},
	};
}

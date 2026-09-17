import { test as base } from '@playwright/test';
import { startIsolatedServer } from './isolated-server.js';

// Overrides the built-in baseURL fixture so every test (Playwright's own
// fixtures are test-scoped by default, so this includes every retry) gets
// its own server instance rather than sharing the one config.ts used to
// start for the whole run -- see isolated-server.ts for why (#168).
export const test = base.extend({
	// biome-ignore lint/correctness/noEmptyPattern: Playwright parses this literal pattern to know which fixtures a fixture depends on -- empty is a real "depends on nothing".
	baseURL: async ({}, use) => {
		const server = await startIsolatedServer();

		await use(server.baseURL);

		server.stop();
	},
});

export { expect } from '@playwright/test';

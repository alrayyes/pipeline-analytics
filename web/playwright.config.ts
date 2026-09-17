import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './tests/e2e',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	reporter: process.env.CI ? [['github'], ['html', { open: 'never' }]] : 'list',
	// Builds the frontend and the Go binary once; each test then starts its
	// own server from that binary via the baseURL fixture in fixtures.ts,
	// rather than the whole run sharing one server process (#168).
	globalSetup: './tests/e2e/global-setup.ts',
	use: {
		trace: 'retain-on-failure',
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});

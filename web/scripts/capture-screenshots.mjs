// Captures the README's screenshots against a real running server, with
// representative mock API responses standing in for real pipeline history
// -- same route-mocking approach as tests/e2e/dashboard.spec.ts, reused
// here for documentation rather than assertions. Run via
// capture-screenshots.sh, which builds the frontend/binary and starts the
// server this script points at.
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';

const BASE_URL = 'http://localhost:4190';
const OUT_DIR = fileURLToPath(new URL('../../docs/', import.meta.url));

const browser = await chromium.launch();
const context = await browser.newContext({
	viewport: { width: 900, height: 520 },
});
const page = await context.newPage();

const cdp = await context.newCDPSession(page);
await cdp.send('WebAuthn.enable');
await cdp.send('WebAuthn.addVirtualAuthenticator', {
	options: {
		protocol: 'ctap2',
		transport: 'internal',
		hasResidentKey: true,
		hasUserVerification: true,
		isUserVerified: true,
		automaticPresenceSimulation: true,
	},
});

await page.goto(BASE_URL);
await page.getByRole('button', { name: 'Register your passkey' }).click();
await page.waitForURL(BASE_URL + '/');

// -- Overview --
// The trailing `*` matters: the real fetch is `/api/pipelines?${params}`
// (see +page.svelte), and a glob with no wildcard after the literal path
// only matches a URL that ends exactly there -- a query string means it
// never matches, the mock never intercepts, and the real (empty) backend
// answers instead. Confirmed live: the Go server's own access log showed
// a real /api/pipelines request reaching it after this route was
// registered, meaning nothing was mocking it at all.
// Envelope shape, not a bare array -- the frontend reads `body.pipelines`
// and `body.hasMore` (PipelineListResponse). A bare array left
// `body.pipelines` undefined and crashed downstream with "Cannot read
// properties of undefined (reading 'length')", confirmed live via a
// pageerror listener once the glob fix above stopped masking it.
await page.route('**/api/pipelines*', (route) =>
	route.fulfill({
		json: {
			pipelines: [
				{ id: 'ci', repoId: 'repo-1', name: 'CI', healthStatus: 'healthy' },
				{
					id: 'deploy',
					repoId: 'repo-1',
					name: 'Deploy',
					healthStatus: 'unhealthy',
					triggeredSignals: ['duration_regression', 'flaky_step'],
				},
				{
					id: 'nightly',
					repoId: 'repo-1',
					name: 'Nightly',
					healthStatus: 'healthy',
				},
			],
			hasMore: false,
		},
	}),
);
await page.reload();
// #101 defaults the Pipelines list to unhealthy-only, which hides the
// mocked healthy "CI"/"Nightly" pipelines below -- click through to the
// full list the overview screenshot is meant to show. Scoped to the
// health-status radiogroup: ForgeFilter has its own "All" radio too.
await page
	.getByRole('radiogroup', { name: 'Filter by health status' })
	.getByRole('radio', { name: 'All' })
	.click();
await page.waitForSelector('text=CI');
await page.screenshot({ path: OUT_DIR + 'screenshot-overview.png' });

// -- Pipeline detail (trend charts) --
const timestamps = Array.from({ length: 6 }, (_, i) =>
	new Date(Date.UTC(2026, 8, 1 + i)).toISOString(),
);
await page.route('**/api/pipelines/deploy', (route) =>
	route.fulfill({
		json: {
			id: 'deploy',
			repoId: 'repo-1',
			name: 'Deploy',
			healthStatus: 'unhealthy',
			triggeredSignals: ['duration_regression', 'failure_rate'],
			durationTrend: {
				timestamps,
				p50: [300, 305, 295, 300, 600, 650],
				p90: [420, 430, 410, 425, 900, 980],
			},
			failureRateTrend: {
				timestamps,
				rate: [0.05, 0.1, 0.05, 0.1, 0.3, 0.35],
			},
		},
	}),
);
await page.route('**/api/pipelines/deploy/steps', (route) =>
	route.fulfill({
		json: [
			{
				id: 'step-1',
				name: 'run tests',
				durationContributionSeconds: 420,
				queueSeconds: 30,
				execSeconds: 390,
				failureRate: 0,
				flaky: false,
			},
			{
				id: 'step-2',
				name: 'flaky integration test',
				durationContributionSeconds: 120,
				queueSeconds: 5,
				execSeconds: 115,
				failureRate: 0.3,
				flaky: true,
				forgeUrl:
					'https://github.com/alrayyes/pipeline-analytics/actions/runs/1/job/2',
			},
			{
				id: 'step-3',
				name: 'deploy to prod',
				durationContributionSeconds: 60,
				queueSeconds: 2,
				execSeconds: 58,
				failureRate: 1,
				flaky: false,
			},
		],
	}),
);
await page.goto(BASE_URL + '/pipelines/deploy');
await page.waitForSelector('table');
await page.screenshot({
	path: OUT_DIR + 'screenshot-pipeline-detail.png',
	fullPage: true,
});

// -- Usage view --
await page.route('**/api/repos/repo-1/usage', (route) =>
	route.fulfill({
		json: [
			{ workflow: 'CI', runnerMinutes: 842.5 },
			{ workflow: 'Deploy', runnerMinutes: 120.2 },
			{ workflow: 'Nightly', runnerMinutes: 30 },
		],
	}),
);
await page.goto(BASE_URL + '/repos/repo-1/usage');
await page.waitForSelector('table');
await page.screenshot({ path: OUT_DIR + 'screenshot-usage.png' });

// -- Dark mode (overview + pipeline detail: cards/badges and charts/table,
// the two most visually distinct surfaces) --
await page.evaluate(() => localStorage.setItem('theme', 'dark'));
await page.goto(BASE_URL + '/');
// Fresh navigation, so showAll (#101) is back to its unhealthy-only default.
await page
	.getByRole('radiogroup', { name: 'Filter by health status' })
	.getByRole('radio', { name: 'All' })
	.click();
await page.waitForSelector('text=CI');
await page.screenshot({ path: OUT_DIR + 'screenshot-overview-dark.png' });

await page.goto(BASE_URL + '/pipelines/deploy');
await page.waitForSelector('table');
await page.screenshot({
	path: OUT_DIR + 'screenshot-pipeline-detail-dark.png',
	fullPage: true,
});

console.log('Screenshots captured.');
await browser.close();

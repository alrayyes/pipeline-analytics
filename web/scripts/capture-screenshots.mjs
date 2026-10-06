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

// -- Pipelines overview --
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
// The Pipelines list lives at /pipelines; / is the failure overview.
await page.goto(BASE_URL + '/pipelines');
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

// -- Telemetry views (failure overview, runs, root cause, flaky) --
// A registered repo, so the views show data rather than the "register a
// repository" empty state.
await page.route('**/api/repos*', (route) =>
	route.fulfill({
		json: {
			repos: [
				{ id: 'repo-1', forge: 'github', identifier: 'alrayyes/payments' },
			],
			hasMore: false,
		},
	}),
);
await page.route('**/api/insights/failures*', (route) =>
	route.fulfill({
		json: {
			window: '7d',
			totalRuns: 658,
			failedRuns: 142,
			passRate: 0.784,
			passRateDelta: -4.2,
			flakyStepRatio: 0.068,
			mttrSeconds: 2280,
			stageDistribution: [
				{ step: 'Run unit tests', failures: 65, share: 0.46 },
				{ step: 'Docker build', failures: 34, share: 0.24 },
				{ step: 'Publish release', failures: 43, share: 0.3 },
			],
			categoryBreakdown: [
				{ category: 'code_tests', occurrences: 65, share: 0.46 },
				{ category: 'config_secrets', occurrences: 34, share: 0.24 },
				{ category: 'network_timeouts', occurrences: 43, share: 0.3 },
			],
			topFailingPipelines: [
				{
					pipelineId: 'deploy',
					pipelineName: 'Deploy',
					repoId: 'repo-1',
					runs: 100,
					failedRuns: 34,
				},
				{
					pipelineId: 'ci',
					pipelineName: 'CI',
					repoId: 'repo-1',
					runs: 400,
					failedRuns: 65,
				},
			],
			failureGroups: [
				{
					step: 'Run unit tests',
					category: 'code_tests',
					conclusion: 'failure',
					occurrences: 65,
					pipelines: [{ pipelineId: 'ci', pipelineName: 'CI' }],
				},
				{
					step: 'Publish release',
					category: 'network_timeouts',
					conclusion: 'timed_out',
					occurrences: 43,
					pipelines: [{ pipelineId: 'deploy', pipelineName: 'Deploy' }],
				},
			],
		},
	}),
);
await page.goto(BASE_URL + '/');
await page.waitForSelector('text=Run unit tests');
await page.screenshot({ path: OUT_DIR + 'screenshot-failure-overview.png' });

await page.goto(BASE_URL + '/failures');
await page.waitForSelector('text=Failing steps');
await page.screenshot({
	path: OUT_DIR + 'screenshot-root-cause.png',
	fullPage: true,
});

const runSteps = (outcomes) =>
	outcomes.map(([name, outcome]) => ({
		name,
		status: outcome === 'queued' ? 'queued' : 'completed',
		outcome,
	}));
await page.route('**/api/runs*', (route) =>
	route.fulfill({
		json: {
			hasMore: false,
			runs: [
				{
					id: 'run-1',
					pipelineId: 'deploy',
					pipelineName: 'Deploy',
					repoId: 'repo-1',
					status: 'completed',
					conclusion: 'failure',
					outcome: 'failed',
					actions: ['rerun'],
					startedAt: '2026-10-02T12:00:00Z',
					durationSeconds: 402,
					branch: 'main',
					sha: 'c4d291a0e2b34f56a7c8d9e0f1a2b3c4d5e6f7a8',
					message: 'fix(stripe): webhook retry',
					actor: 'marcus-v',
					forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/1',
					steps: runSteps([
						['checkout', 'passed'],
						['build', 'passed'],
						['canary', 'failed'],
						['promote', 'queued'],
					]),
				},
				{
					id: 'run-2',
					pipelineId: 'ci',
					pipelineName: 'CI',
					repoId: 'repo-1',
					status: 'completed',
					conclusion: 'success',
					outcome: 'passed',
					actions: ['rerun'],
					startedAt: '2026-10-02T11:00:00Z',
					durationSeconds: 115,
					branch: 'main',
					sha: 'a1b2c3d4e5f60718293a4b5c6d7e8f9012345678',
					message: 'chore: bump dependencies',
					actor: 'renovate',
					forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/2',
					steps: runSteps([
						['checkout', 'passed'],
						['test', 'passed'],
						['lint', 'passed'],
					]),
				},
			],
		},
	}),
);
await page.goto(BASE_URL + '/runs');
await page.waitForSelector('text=Deploy');
await page.screenshot({ path: OUT_DIR + 'screenshot-runs.png' });

const flakyHistory = (failedAt, length) =>
	Array.from({ length }, (_, i) =>
		failedAt.includes(i) ? 'failed' : 'passed',
	);
await page.route('**/api/steps/flaky*', (route) =>
	route.fulfill({
		json: {
			window: '7d',
			hasMore: false,
			steps: [
				{
					pipelineId: 'ci',
					pipelineName: 'CI',
					repoId: 'repo-1',
					name: 'Run unit tests',
					flakeRate: 0.25,
					runCount: 120,
					recentOutcomes: flakyHistory([3, 17, 30, 39], 40),
				},
				{
					pipelineId: 'deploy',
					pipelineName: 'Deploy',
					repoId: 'repo-1',
					name: 'Browser tests',
					flakeRate: 0.1,
					runCount: 40,
					recentOutcomes: flakyHistory([8, 22], 40),
				},
			],
		},
	}),
);
await page.goto(BASE_URL + '/flaky');
await page.waitForSelector('text=Browser tests');
await page.screenshot({ path: OUT_DIR + 'screenshot-flaky.png' });

// -- Dark mode (overview + pipeline detail: cards/badges and charts/table,
// the two most visually distinct surfaces) --
await page.evaluate(() => localStorage.setItem('theme', 'dark'));
await page.goto(BASE_URL + '/pipelines');
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

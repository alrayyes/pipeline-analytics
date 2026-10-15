import type { BrowserContext, Page, Route } from '@playwright/test';
import { expectFitsAndPassesAxeInBothSchemes } from './axe.js';
import { expect, test } from './fixtures.js';

// A CDP virtual authenticator stands in for a passkey device, as in
// dashboard.spec.ts, then the first-run registration signs the test in.
async function signIn(page: Page, context: BrowserContext): Promise<void> {
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

	await page.goto('/');
	await page.getByRole('button', { name: 'Register your passkey' }).click();
	await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible();
}

const insights = {
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
	],
	categoryBreakdown: [],
	topFailingPipelines: [
		{
			pipelineId: 'p-api',
			pipelineName: 'api-gateway',
			repoId: 'r1',
			runs: 100,
			failedRuns: 34,
		},
		{
			pipelineId: 'p-auth',
			pipelineName: 'auth-service',
			repoId: 'r1',
			runs: 50,
			failedRuns: 14,
		},
	],
	failureGroups: [],
};

const latestFailure = {
	id: 'run-9',
	pipelineId: 'p-deploy',
	pipelineName: 'staging-deploy',
	repoId: 'r1',
	status: 'completed',
	conclusion: 'failure',
	outcome: 'failed',
	startedAt: '2026-10-02T12:00:00Z',
	sha: 'c4d291a0e2b34f56a7c8d9e0f1a2b3c4d5e6f7a8',
	message: 'fix(stripe): webhook retry',
	steps: [],
};

async function mockInsights(
	page: Page,
	respond: (url: URL) => unknown | number,
): Promise<URL[]> {
	const requested: URL[] = [];

	await page.route(/\/api\/insights\/failures/, (route: Route) => {
		const url = new URL(route.request().url());
		requested.push(url);

		const body = respond(url);
		if (typeof body === 'number') return route.fulfill({ status: body });

		return route.fulfill({ json: body });
	});

	return requested;
}

async function mockLatestFailure(page: Page, runs: unknown[]): Promise<void> {
	await page.route(/\/api\/runs\?/, (route) =>
		route.fulfill({ json: { runs, hasMore: false } }),
	);
}

async function mockOneRepo(page: Page): Promise<void> {
	await page.route('**/api/repos*', (route) =>
		route.fulfill({
			json: {
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			},
		}),
	);
}

test('shows the figures, the failing steps, the top failing pipelines and the latest failure', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	const requested = await mockInsights(page, () => insights);
	await mockLatestFailure(page, [latestFailure]);

	await page.goto('/');
	await expect(page).toHaveURL('/');
	// The account's default window, from the real server's settings, is what the
	// first request asks for: the one hop from the settings to the page that
	// only a real server proves.
	await expect.poll(() => requested[0]?.searchParams.get('window')).toBe('7d');
	await expect(
		page.getByRole('heading', { name: 'Overview', level: 1 }),
	).toBeVisible();

	const passRate = page.getByRole('group', { name: 'Pass rate' });
	await expect(passRate.getByText('78%')).toBeVisible();
	// The server's change in percentage points, formatted.
	await expect(passRate.getByText('-4.2 pts')).toBeVisible();

	await expect(
		page
			.getByRole('group', { name: 'Mean time to recovery' })
			.getByText('38.0m'),
	).toBeVisible();
	await expect(
		page.getByRole('group', { name: 'Failed runs' }).getByText('142 / 658'),
	).toBeVisible();
	await expect(
		page.getByRole('group', { name: 'Flaky steps' }).getByText('7%'),
	).toBeVisible();

	const steps = page.getByRole('region', {
		name: 'Failure stage distribution',
	});
	await expect(steps.getByText('Run unit tests')).toBeVisible();
	await expect(steps.getByText('46%')).toBeVisible();
	await expect(steps.getByText('65 failures')).toBeVisible();

	const top = page.getByRole('region', { name: 'Top failing pipelines' });
	await expect(top.getByRole('link', { name: /api-gateway/ })).toHaveAttribute(
		'href',
		'/pipelines/p-api',
	);
	await expect(top.getByText('34 of 100 runs failed')).toBeVisible();

	const banner = page.getByRole('region', { name: 'Latest failure' });
	await expect(banner.getByText('staging-deploy')).toBeVisible();
	await expect(banner.getByText('c4d291a', { exact: true })).toBeVisible();
	await expect(banner.getByRole('link', { name: /inspect/i })).toHaveAttribute(
		'href',
		'/runs/run-9',
	);
});

test('fits a phone and passes axe in light and dark mode', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await page.setViewportSize({ width: 390, height: 844 });
	await mockInsights(page, () => insights);
	await mockLatestFailure(page, [latestFailure]);

	await page.goto('/');
	await expect(page.getByRole('group', { name: 'Pass rate' })).toBeVisible();

	await expectFitsAndPassesAxeInBothSchemes(page, async () => {
		await expect(page.getByRole('group', { name: 'Pass rate' })).toBeVisible();
	});
});

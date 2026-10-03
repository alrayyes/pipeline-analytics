import AxeBuilder from '@axe-core/playwright';
import type { BrowserContext, Page, Route } from '@playwright/test';
import { expect, test } from './fixtures.js';

const a11yTags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

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

test('with no repository registered, the landing page says how to start', async ({
	page,
	context,
}) => {
	await signIn(page, context);

	await expect(page).toHaveURL('/');
	await expect(page.getByText('No repositories registered yet')).toBeVisible();
});

test('shows the figures, the failing steps, the top failing pipelines and the latest failure', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockInsights(page, () => insights);
	await mockLatestFailure(page, [latestFailure]);

	await page.goto('/');
	await expect(page).toHaveURL('/');
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

test('a figure the server has no data for says so instead of showing zero', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockInsights(page, () => ({
		...insights,
		passRate: undefined,
		passRateDelta: undefined,
		mttrSeconds: undefined,
		totalRuns: 3,
		failedRuns: 0,
	}));
	await mockLatestFailure(page, []);

	await page.goto('/');

	await expect(
		page.getByRole('group', { name: 'Pass rate' }).getByText('No data'),
	).toBeVisible();
	await expect(
		page
			.getByRole('group', { name: 'Mean time to recovery' })
			.getByText('No recoveries'),
	).toBeVisible();
	await expect(
		page.getByRole('region', { name: 'Latest failure' }),
	).toHaveCount(0);
});

test('the window toggle asks the server for that window', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	const requested = await mockInsights(page, () => insights);
	await mockLatestFailure(page, []);

	await page.goto('/');
	await expect(page.getByRole('group', { name: 'Pass rate' })).toBeVisible();
	expect(requested[0].searchParams.get('window')).toBe('7d');

	await page.getByRole('radio', { name: '24 hours' }).click();

	await expect
		.poll(() => requested[requested.length - 1].searchParams.get('window'))
		.toBe('24h');
});

test('says so when the figures cannot load', async ({ page, context }) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockInsights(page, () => 500);
	await mockLatestFailure(page, []);

	await page.goto('/');

	await expect(page.getByRole('alert')).toContainText(
		"Couldn't load the overview",
	);
});

for (const scheme of ['light', 'dark'] as const) {
	test(`fits a phone and passes axe in ${scheme} mode`, async ({
		page,
		context,
	}) => {
		await signIn(page, context);
		await mockOneRepo(page);
		await page.setViewportSize({ width: 390, height: 844 });
		await page.emulateMedia({ colorScheme: scheme });
		await mockInsights(page, () => insights);
		await mockLatestFailure(page, [latestFailure]);

		await page.goto('/');
		await expect(page.getByRole('group', { name: 'Pass rate' })).toBeVisible();

		const overflow = await page.evaluate(
			() => document.documentElement.scrollWidth - window.innerWidth,
		);
		expect(overflow).toBeLessThanOrEqual(0);

		const scan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
		expect(scan.violations).toEqual([]);
	});
}

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
	totalRuns: 120,
	failedRuns: 30,
	flakyStepRatio: 0.1,
	stageDistribution: [
		{ step: 'Run unit tests', failures: 14, share: 0.5 },
		{ step: 'Docker login', failures: 8, share: 0.2857 },
		{ step: 'Publish release', failures: 6, share: 0.2143 },
	],
	categoryBreakdown: [
		{ category: 'code_tests', occurrences: 14, share: 0.5 },
		{ category: 'config_secrets', occurrences: 8, share: 0.2857 },
		{ category: 'uncategorised', occurrences: 6, share: 0.2143 },
	],
	topFailingPipelines: [],
	failureGroups: [
		{
			step: 'Run unit tests',
			category: 'code_tests',
			conclusion: 'failure',
			occurrences: 14,
			pipelines: [
				{ pipelineId: 'p-api', pipelineName: 'api-gateway' },
				{ pipelineId: 'p-web', pipelineName: 'web-client' },
			],
		},
		{
			step: 'Docker login',
			category: 'config_secrets',
			conclusion: 'failure',
			occurrences: 8,
			pipelines: [{ pipelineId: 'p-api', pipelineName: 'api-gateway' }],
		},
		{
			step: 'Publish release',
			category: 'uncategorised',
			conclusion: 'timed_out',
			occurrences: 6,
			pipelines: [{ pipelineId: 'p-rel', pipelineName: 'release' }],
		},
	],
};

const empty = {
	totalRuns: 10,
	failedRuns: 0,
	flakyStepRatio: 0,
	stageDistribution: [],
	categoryBreakdown: [],
	topFailingPipelines: [],
	failureGroups: [],
};

// Answers GET /api/insights/failures from `respond`, recording every
// requested URL so a test can assert on the window the page asked for.
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

async function mockOneRepo(page: Page): Promise<void> {
	await page.route('**/api/repos*', (route) =>
		route.fulfill({
			json: {
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			},
		}),
	);
}

test('shows the category breakdown and each failing step, with links to its failed runs', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockInsights(page, () => insights);

	await page.goto('/');
	await page.getByRole('link', { name: 'Failures', exact: true }).click();
	await expect(page).toHaveURL('/failures');
	await expect(
		page.getByRole('heading', { name: 'Root cause diagnostics', level: 1 }),
	).toBeVisible();

	const categories = page.getByRole('region', { name: 'Failure categories' });
	await expect(categories.getByText('Code and tests')).toBeVisible();
	await expect(categories.getByText('50%')).toBeVisible();
	await expect(categories.getByText('14 failures')).toBeVisible();
	await expect(categories.getByText('Config and secrets')).toBeVisible();
	// The server's share, formatted: 0.2857 reads 29%.
	await expect(categories.getByText('29%')).toBeVisible();

	const tests = page
		.getByRole('listitem')
		.filter({ hasText: 'Run unit tests' });
	await expect(tests.getByText('Code and tests')).toBeVisible();
	await expect(tests.getByText('14 failures in 2 pipelines')).toBeVisible();
	await expect(
		tests.getByRole('link', { name: /api-gateway/ }),
	).toHaveAttribute(
		'href',
		'/pipelines/p-api/flaky-runs?step=Run%20unit%20tests',
	);
	await expect(tests.getByRole('link', { name: /web-client/ })).toHaveAttribute(
		'href',
		'/pipelines/p-web/flaky-runs?step=Run%20unit%20tests',
	);

	// The raw conclusion stays visible next to the heuristic category, so the
	// guess can be checked against what the forge said.
	const publish = page
		.getByRole('listitem')
		.filter({ hasText: 'Publish release' });
	await expect(publish.getByText('timed_out')).toBeVisible();
	await expect(publish.getByText('1 pipeline', { exact: false })).toBeVisible();
});

test('the window toggle asks the server for that window', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	const requested = await mockInsights(page, () => insights);

	await page.goto('/failures');
	await expect(page.getByText('Run unit tests').first()).toBeVisible();
	// The default is 7d, the account's default window.
	expect(requested[0].searchParams.get('window')).toBe('7d');

	await page.getByRole('radio', { name: '30 days' }).click();

	await expect
		.poll(() => requested[requested.length - 1].searchParams.get('window'))
		.toBe('30d');
	await expect(page.getByRole('radio', { name: '30 days' })).toBeChecked();
});

test('says so when nothing failed, and when the figures cannot load', async ({
	page,
	context,
}) => {
	await signIn(page, context);

	let respond: unknown | number = empty;
	await mockInsights(page, () => respond);

	await page.goto('/failures');
	await expect(page.getByText('No failures in this window.')).toBeVisible();
	await expect(
		page.getByRole('region', { name: 'Failure categories' }),
	).toHaveCount(0);

	respond = 500;
	await page.getByRole('radio', { name: '24 hours' }).click();
	await expect(page.getByRole('alert')).toContainText("Couldn't load failures");
});

for (const scheme of ['light', 'dark'] as const) {
	test(`fits a phone and passes axe in ${scheme} mode`, async ({
		page,
		context,
	}) => {
		await signIn(page, context);
		await page.setViewportSize({ width: 390, height: 844 });
		await page.emulateMedia({ colorScheme: scheme });
		await mockInsights(page, () => insights);

		await page.goto('/failures');
		await expect(page.getByText('Run unit tests').first()).toBeVisible();

		const overflow = await page.evaluate(
			() => document.documentElement.scrollWidth - window.innerWidth,
		);
		expect(overflow).toBeLessThanOrEqual(0);

		const scan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
		expect(scan.violations).toEqual([]);
	});
}

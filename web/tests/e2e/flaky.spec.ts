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

const history = (failedAt: number[], length: number): string[] =>
	Array.from({ length }, (_, i) =>
		failedAt.includes(i) ? 'failed' : 'passed',
	);

const flaky = {
	window: '7d',
	steps: [
		{
			pipelineId: 'p-api',
			pipelineName: 'api-gateway',
			repoId: 'r1',
			name: 'Run unit tests',
			flakeRate: 0.25,
			runCount: 120,
			recentOutcomes: history([3, 17, 30, 39], 40),
			quarantined: false,
		},
		{
			pipelineId: 'p-web',
			pipelineName: 'web-client',
			repoId: 'r1',
			name: 'Browser tests',
			flakeRate: 0.1,
			runCount: 5,
			recentOutcomes: ['passed', 'failed', 'passed', 'passed', 'passed'],
			quarantined: false,
		},
	],
	hasMore: false,
};

// The same two steps with the second one quarantined two days ago, for 28
// more days (quarantine-flaky-steps: still flaky, still listed).
const day = 86_400_000;
const withQuarantine = {
	...flaky,
	steps: [
		flaky.steps[0],
		{
			...flaky.steps[1],
			quarantined: true,
			quarantine: {
				note: 'Waiting on the vendor fix',
				quarantinedAt: new Date(Date.now() - 2 * day).toISOString(),
				expiresAt: new Date(Date.now() + 28 * day).toISOString(),
			},
		},
	],
};

const empty = { window: '7d', steps: [], hasMore: false };

// Answers GET /api/steps/flaky from `respond`, recording every requested URL
// so a test can assert on the window the page asked for.
async function mockFlaky(
	page: Page,
	respond: (url: URL) => unknown | number,
): Promise<URL[]> {
	const requested: URL[] = [];

	await page.route(/\/api\/steps\/flaky/, (route: Route) => {
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

test('shows each flaky step with its matrix, rate and run count, flakiest first', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockFlaky(page, () => flaky);

	await page.goto('/');
	await page.getByRole('link', { name: 'Flaky', exact: true }).click();
	await expect(page).toHaveURL('/flaky');
	await expect(
		page.getByRole('heading', { name: 'Flaky tests', level: 1 }),
	).toBeVisible();

	const items = page
		.getByRole('listitem')
		.filter({ has: page.getByRole('img') });
	await expect(items).toHaveCount(2);
	await expect(items.first()).toContainText('Run unit tests');

	const unit = items.filter({ hasText: 'Run unit tests' });
	await expect(unit.getByText('api-gateway')).toBeVisible();
	await expect(unit.getByText('Flaky', { exact: true })).toBeVisible();
	await expect(
		unit
			.getByText(
				'4 of the last 40 runs failed. 25% failure rate over 120 runs.',
			)
			.first(),
	).toBeVisible();

	// The matrix has a text equivalent: readable without seeing the grid.
	await expect(
		unit.getByRole('img', {
			name: '4 of the last 40 runs failed. 25% failure rate over 120 runs.',
		}),
	).toBeVisible();

	// A step with fewer than 40 runs shows only what it has.
	const browser = items.filter({ hasText: 'Browser tests' });
	await expect(
		browser.getByRole('img', {
			name: '1 of the last 5 runs failed. 10% failure rate over 5 runs.',
		}),
	).toBeVisible();

	await expect(
		unit.getByRole('link', { name: /failed runs/i }),
	).toHaveAttribute(
		'href',
		'/pipelines/p-api/flaky-runs?step=Run%20unit%20tests',
	);
});

test('the window toggle asks the server for that window', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	const requested = await mockFlaky(page, () => flaky);

	await page.goto('/flaky');
	await expect(page.getByText('Run unit tests').first()).toBeVisible();
	expect(requested[0].searchParams.get('window')).toBe('7d');

	await page.getByRole('radio', { name: '30 days' }).click();

	await expect
		.poll(() => requested[requested.length - 1].searchParams.get('window'))
		.toBe('30d');
	await expect(page.getByRole('radio', { name: '30 days' })).toBeChecked();
});

test('says so when nothing is flaky, and when the list cannot load', async ({
	page,
	context,
}) => {
	await signIn(page, context);

	let respond: unknown | number = empty;
	await mockFlaky(page, () => respond);

	await page.goto('/flaky');
	await expect(page.getByText('No flaky steps in this window.')).toBeVisible();

	respond = 500;
	await page.getByRole('radio', { name: '24 hours' }).click();
	await expect(page.getByRole('alert')).toContainText(
		"Couldn't load flaky steps",
	);
});

test('with no saved window it asks for none, and shows the one the server used', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.evaluate(() => localStorage.clear());
	await page.route('**/api/settings', (route) =>
		route.fulfill({ status: 500 }),
	);
	const requested = await mockFlaky(page, () => ({ ...flaky, window: '24h' }));

	await page.goto('/flaky');

	await expect(page.getByRole('radio', { name: '24 hours' })).toBeChecked();
	expect(requested[0].searchParams.has('window')).toBe(false);
});

for (const scheme of ['light', 'dark'] as const) {
	test(`fits a phone and passes axe in ${scheme} mode`, async ({
		page,
		context,
	}) => {
		await signIn(page, context);
		await page.setViewportSize({ width: 390, height: 844 });
		await page.emulateMedia({ colorScheme: scheme });
		await mockFlaky(page, () => withQuarantine);

		await page.goto('/flaky');
		await expect(page.getByText('Run unit tests').first()).toBeVisible();
		await expect(page.getByText('Quarantined', { exact: true })).toBeVisible();

		const overflow = await page.evaluate(
			() => document.documentElement.scrollWidth - window.innerWidth,
		);
		expect(overflow).toBeLessThanOrEqual(0);

		const scan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
		expect(scan.violations).toEqual([]);
	});
}

test('quarantines a flaky step with a note, and un-quarantines it again', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockFlaky(page, () => withQuarantine);

	const writes: { method: string; url: string; body: string | null }[] = [];
	await page.route(
		/\/api\/pipelines\/[^/]+\/steps\/[^/]+\/quarantine$/,
		(route) => {
			const request = route.request();
			writes.push({
				method: request.method(),
				url: new URL(request.url()).pathname,
				body: request.postData(),
			});

			if (request.method() === 'DELETE') {
				return route.fulfill({ json: { quarantined: false } });
			}

			return route.fulfill({
				json: {
					quarantined: true,
					quarantine: {
						note: 'Race in the lock',
						quarantinedAt: new Date().toISOString(),
						expiresAt: new Date(Date.now() + 30 * day).toISOString(),
					},
				},
			});
		},
	);

	await page.goto('/flaky');

	const items = page
		.getByRole('listitem')
		.filter({ has: page.getByRole('img') });
	const unit = items.filter({ hasText: 'Run unit tests' });
	const browser = items.filter({ hasText: 'Browser tests' });

	// A quarantined step says so in words, with its age, note and expiry, and
	// is still listed as flaky.
	await expect(browser.getByText('Quarantined', { exact: true })).toBeVisible();
	await expect(browser.getByText('Flaky', { exact: true })).toBeVisible();
	await expect(browser.getByText('2 days ago')).toBeVisible();
	await expect(browser.getByText('in 4 weeks')).toBeVisible();
	await expect(browser.getByText('Waiting on the vendor fix')).toBeVisible();
	await expect(unit.getByText('Quarantined', { exact: true })).toHaveCount(0);

	// Quarantine from the list, with a note: no reload, the label appears.
	await unit.getByRole('button', { name: 'Quarantine', exact: true }).click();
	await unit.getByLabel('Note (optional)').fill('Race in the lock');
	await unit.getByRole('button', { name: 'Quarantine this step' }).click();

	await expect(unit.getByText('Quarantined', { exact: true })).toBeVisible();
	await expect(unit.getByText('Race in the lock')).toBeVisible();
	await expect(
		unit.getByRole('button', { name: 'Un-quarantine' }),
	).toBeVisible();
	expect(writes[0]).toEqual({
		method: 'PUT',
		url: '/api/pipelines/p-api/steps/Run%20unit%20tests/quarantine',
		body: '{"note":"Race in the lock"}',
	});

	const scan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
	expect(scan.violations).toEqual([]);

	// And back out again.
	await browser.getByRole('button', { name: 'Un-quarantine' }).click();
	await expect(browser.getByText('Quarantined', { exact: true })).toHaveCount(
		0,
	);
	await expect(browser.getByText('Flaky', { exact: true })).toBeVisible();
	expect(writes[1]).toEqual({
		method: 'DELETE',
		url: '/api/pipelines/p-web/steps/Browser%20tests/quarantine',
		body: null,
	});
});

test('says so when the quarantine could not be saved', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockFlaky(page, () => flaky);
	await page.route(/\/quarantine$/, (route) => route.fulfill({ status: 500 }));

	await page.goto('/flaky');
	const unit = page.getByRole('listitem').filter({ hasText: 'Run unit tests' });
	await unit.getByRole('button', { name: 'Quarantine', exact: true }).click();
	await unit.getByRole('button', { name: 'Quarantine this step' }).click();

	await expect(unit.getByRole('alert')).toContainText("Couldn't update");
	await expect(unit.getByText('Quarantined', { exact: true })).toHaveCount(0);
});

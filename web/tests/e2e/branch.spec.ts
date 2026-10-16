import type { BrowserContext, Page } from '@playwright/test';
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
	totalRuns: 40,
	failedRuns: 10,
	passRate: 0.75,
	flakyStepRatio: 0.1,
	stageDistribution: [],
	categoryBreakdown: [],
	topFailingPipelines: [],
	failureGroups: [],
};

// The server answers every read from the branch it was asked for: main has
// runs, the other branch has none in the window.
async function mockTelemetry(
	page: Page,
	requested: URL[],
	requestedRuns: URL[] = [],
): Promise<void> {
	await page.route('**/api/repos*', (route) =>
		route.fulfill({
			json: {
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			},
		}),
	);
	await page.route('**/api/branches*', (route) =>
		route.fulfill({
			json: {
				window: '7d',
				branches: [
					{ name: 'main', runCount: 30 },
					{ name: 'feat/quiet', runCount: 1 },
				],
			},
		}),
	);
	await page.route('**/api/insights/failures*', (route) => {
		const url = new URL(route.request().url());
		requested.push(url);

		return route.fulfill({
			json:
				url.searchParams.get('branch') === 'feat/quiet'
					? { ...insights, totalRuns: 0, failedRuns: 0, passRate: undefined }
					: insights,
		});
	});
	await page.route('**/api/runs*', (route) => {
		requestedRuns.push(new URL(route.request().url()));

		return route.fulfill({ json: { runs: [], hasMore: false } });
	});
	await page.route('**/api/steps/flaky*', (route) => {
		requested.push(new URL(route.request().url()));

		return route.fulfill({
			json: { window: '7d', steps: [], hasMore: false },
		});
	});
}

test('picking a branch scopes the views, and a branch with no runs says so', async ({
	page,
	context,
}) => {
	const requested: URL[] = [];
	const requestedRuns: URL[] = [];
	await signIn(page, context);
	await mockTelemetry(page, requested, requestedRuns);

	await page.goto('/');
	await expect(page.getByText('Pass rate')).toBeVisible();
	await expect(page.getByLabel('Branch')).toContainText('All branches');
	expect(requested.at(-1)?.searchParams.has('branch')).toBe(false);

	await page.getByLabel('Branch').click();
	await page.getByRole('option', { name: /^main/ }).click();
	await expect(page.getByLabel('Branch')).toContainText('main');
	await expect
		.poll(() => requested.at(-1)?.searchParams.get('branch'))
		.toBe('main');
	// The latest-failure card is part of the overview, so it follows the branch.
	await expect
		.poll(() => requestedRuns.at(-1)?.searchParams.get('branch'))
		.toBe('main');

	// The choice carries to the next view, through the nav rather than a
	// reload: it lasts the visit, not the page load.
	await page.getByRole('link', { name: 'Flaky' }).first().click();
	await expect(page.getByLabel('Branch')).toContainText('main');
	await expect
		.poll(() => requested.at(-1)?.searchParams.get('branch'))
		.toBe('main');

	await page.getByLabel('Branch').click();
	await page.getByRole('option', { name: /^feat\/quiet/ }).click();
	await expect(
		page.getByText('No flaky steps on feat/quiet in this window.'),
	).toBeVisible();

	await page.getByLabel('Branch').click();
	await page.getByRole('option', { name: 'All branches' }).click();
	await expect(page.getByText('No flaky steps in this window.')).toBeVisible();
});

test('the selector fits a phone and passes axe in light and dark mode', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.setViewportSize({ width: 390, height: 844 });
	await mockTelemetry(page, []);

	await page.goto('/');
	await expect(page.getByText('Pass rate')).toBeVisible();

	await expectFitsAndPassesAxeInBothSchemes(page, async () => {
		await expect(page.getByText('Pass rate')).toBeVisible();
	});
});

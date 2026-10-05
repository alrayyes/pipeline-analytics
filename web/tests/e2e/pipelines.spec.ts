import type { BrowserContext, Page, Route } from '@playwright/test';
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

async function mockOneRepo(page: Page): Promise<void> {
	await page.route('**/api/repos*', (route) =>
		route.fulfill({
			json: {
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			},
		}),
	);
}

interface Fixture {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
}

// Answers GET /api/pipelines as the server does for the health filter: it
// applies across every pipeline, so a filter that matches nothing is an empty
// list even though pipelines exist.
function serve(fixtures: Fixture[]) {
	return (route: Route) => {
		const health = new URL(route.request().url()).searchParams.get('health');

		return route.fulfill({
			json: {
				pipelines: fixtures.filter((p) => !health || p.healthStatus === health),
				hasMore: false,
			},
		});
	};
}

test('a filter that matches nothing says so and keeps the controls, instead of claiming nothing is ingested', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	// Every pipeline is healthy, and the filter defaults to Unhealthy.
	await page.route(
		'**/api/pipelines*',
		serve([{ id: 'ci', repoId: 'r1', name: 'CI', healthStatus: 'healthy' }]),
	);

	await page.goto('/pipelines');

	await expect(
		page.getByText('No pipelines match the selected filters.'),
	).toBeVisible();
	await expect(page.getByText('No pipeline runs ingested yet')).toHaveCount(0);

	const healthFilter = page.getByRole('radiogroup', {
		name: 'Filter by health status',
	});
	await expect(healthFilter).toBeVisible();

	await healthFilter.getByRole('radio', { name: 'All' }).click();
	await expect(page.getByRole('heading', { name: 'CI' })).toBeVisible();
});

test('with no pipelines at all, it still says nothing has been ingested', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await page.route('**/api/pipelines*', serve([]));

	await page.goto('/pipelines');

	await expect(page.getByText('No pipeline runs ingested yet')).toBeVisible();
	await expect(
		page.getByText('No pipelines match the selected filters.'),
	).toHaveCount(0);
});

test('"Reset filters" follows the server\'s defaults, with no copy in the page', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await page.route(
		'**/api/pipelines*',
		serve([{ id: 'ci', repoId: 'r1', name: 'CI', healthStatus: 'unhealthy' }]),
	);

	await page.goto('/pipelines');

	const reset = page.getByRole('button', { name: 'Reset filters' });
	const healthFilter = page.getByRole('radiogroup', {
		name: 'Filter by health status',
	});

	// A fresh account is on the server's defaults, whatever they are.
	await expect(reset).toBeDisabled();

	await healthFilter.getByRole('radio', { name: 'All' }).click();
	await expect(reset).toBeEnabled();

	await reset.click();
	await expect(reset).toBeDisabled();
	await expect(
		healthFilter.getByRole('radio', { name: 'Unhealthy' }),
	).toBeChecked();
});

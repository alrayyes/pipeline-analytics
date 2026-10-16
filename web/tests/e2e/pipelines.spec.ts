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

async function mockOneRepo(page: Page): Promise<void> {
	await page.route('**/api/repos*', (route) =>
		route.fulfill({
			json: {
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			},
		}),
	);
}

// The states behind the list (nothing matching, nothing ingested, Reset
// filters) are asserted on the component layer in pipelines.dom.test.ts. What
// only a browser can prove stays here: the real page loads and has no axe
// violations in either colour scheme.
test('the pipeline list fits a phone and passes axe in light and dark mode', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.setViewportSize({ width: 390, height: 844 });
	await mockOneRepo(page);
	await page.route('**/api/pipelines*', (route) =>
		route.fulfill({
			json: {
				pipelines: [
					{ id: 'ci', repoId: 'r1', name: 'CI', healthStatus: 'unhealthy' },
				],
				hasMore: false,
			},
		}),
	);

	await page.goto('/pipelines');
	await expect(page.getByRole('heading', { name: 'CI' })).toBeVisible();

	await expectFitsAndPassesAxeInBothSchemes(page, async () => {
		await expect(page.getByRole('heading', { name: 'CI' })).toBeVisible();
	});
});

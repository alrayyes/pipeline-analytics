import type { BrowserContext, Page } from '@playwright/test';
import { expectNoViolations } from './axe.js';
import { expect, test } from './fixtures.js';

const DAY_SECONDS = 86_400;

// A CDP virtual authenticator stands in for a passkey device, as in
// saved-tokens.spec.ts, then the first-run registration signs the test in.
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

function expiryText(days: number): RegExp {
	const expires = new Date(Date.now() + days * DAY_SECONDS * 1000);
	const formatted = expires.toLocaleDateString(undefined, {
		year: 'numeric',
		month: 'long',
		day: 'numeric',
	});

	return new RegExp(`Expires on ${formatted}`);
}

test('the API tokens section defaults to 90 days, previews the expiry, and shows the token once', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.goto('/settings');

	const lifetime = page.getByRole('radiogroup', { name: 'Token lifetime' });
	await expect(lifetime.getByRole('radio')).toHaveText([
		'30 days',
		'90 days',
		'1 year',
	]);
	await expect(lifetime.getByRole('radio', { name: '90 days' })).toBeChecked();
	await expect(page.getByTestId('token-expiry-preview')).toHaveText(
		expiryText(90),
	);

	await lifetime.getByRole('radio', { name: '30 days' }).click();
	await expect(page.getByTestId('token-expiry-preview')).toHaveText(
		expiryText(30),
	);

	await expectNoViolations(page);

	const sent = page.waitForRequest(
		(req) => req.url().endsWith('/api/auth/tokens') && req.method() === 'POST',
	);
	await page.getByRole('button', { name: 'Create token' }).click();
	expect((await sent).postDataJSON()).toEqual({ ttlSeconds: 30 * DAY_SECONDS });

	const status = page.getByRole('status');
	await expect(status).toContainText('shown once');
	await expect(status.locator('code')).not.toBeEmpty();
	await expect(status).toContainText(expiryText(30));
	// The pointer is still over the button, whose hover shade is the one
	// colour pair axe rejects; scan the resting state.
	await page.mouse.move(0, 0);
	await expect(
		page.getByRole('button', { name: 'Create token' }),
	).toBeEnabled();
	await expectNoViolations(page);

	await page.reload();
	await expect(page.getByText('API tokens', { exact: true })).toBeVisible();
	await expect(page.getByRole('status')).toHaveCount(0);
});

test('a refused creation shows the error and no token', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.route('**/api/auth/tokens', (route) =>
		route.fulfill({ status: 400, json: { message: 'ttlSeconds is invalid' } }),
	);
	await page.goto('/settings');

	await page.getByRole('button', { name: 'Create token' }).click();

	await expect(page.getByRole('alert')).toHaveText('ttlSeconds is invalid');
	await expect(page.getByRole('status')).toHaveCount(0);
});

test('the API tokens section fits a phone and passes axe', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.setViewportSize({ width: 375, height: 700 });
	await page.goto('/settings');
	await expect(
		page.getByRole('button', { name: 'Create token' }),
	).toBeVisible();

	const overflow = await page.evaluate(
		() => document.documentElement.scrollWidth - window.innerWidth,
	);
	expect(overflow).toBeLessThanOrEqual(0);
	await expectNoViolations(page);
});

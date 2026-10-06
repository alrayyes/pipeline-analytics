import type { BrowserContext, Page } from '@playwright/test';
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

async function saveGitHubToken(page: Page, token: string): Promise<void> {
	await page.goto('/settings');
	await page.getByLabel('Access token').fill(token);
	await page.getByRole('button', { name: 'Save token' }).click();
	await expect(page.getByText(/^Saved GitHub token/)).toBeVisible();
}

test('a saved token is used by the register dialog, and no response carries it', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.route('**/api/repos/discover', (route) => {
		const sent = route.request().postDataJSON();
		expect(sent).not.toHaveProperty('token');

		return route.fulfill({ json: ['alrayyes/demo-repo'] });
	});

	await page.goto('/repos');
	await page.getByRole('button', { name: 'Register repository' }).click();
	await expect(page.getByLabel('Access token')).toBeVisible();
	// Read-only Actions tracks a repo; re-running and cancelling need write.
	await expect(page.getByRole('dialog')).toContainText(
		'Actions (read-only) is enough to track a repository',
	);
	await expect(page.getByRole('dialog')).toContainText(
		'Actions write permission is also needed to re-run or cancel runs',
	);
	await page.keyboard.press('Escape');

	await saveGitHubToken(page, 'ghp_secrettoken9999');
	await expect(page.getByRole('row', { name: /GitHub/ })).toContainText(
		'****9999',
	);

	// Nothing the API says contains the token, only its masked form.
	const listed = await page.request.get('/api/forge-tokens');
	expect(await listed.text()).not.toContain('secrettoken');

	// A restart of the browser is a reload with its storage gone.
	await page.evaluate(() => localStorage.clear());
	await page.goto('/repos');
	await page.getByRole('button', { name: 'Register repository' }).click();

	await expect(
		page.getByRole('heading', { name: 'Select repositories to follow' }),
	).toBeVisible();
	await expect(page.getByLabel('Access token')).toHaveCount(0);
	await expect(page.getByText('GitHub · saved token ****9999')).toBeVisible();
});

test('saving for the same forge replaces the token, and deleting asks for one again', async ({
	page,
	context,
}) => {
	await signIn(page, context);

	await saveGitHubToken(page, 'ghp_first1111');
	await saveGitHubToken(page, 'ghp_second2222');

	const rows = page.getByRole('row', { name: /GitHub/ });
	await expect(rows).toHaveCount(1);
	await expect(rows).toContainText('****2222');

	await rows.getByRole('button', { name: /Delete/ }).click();
	await page
		.getByRole('alertdialog')
		.getByRole('button', { name: 'Delete' })
		.click();
	await expect(page.getByText('No saved tokens.')).toBeVisible();

	await page.goto('/repos');
	await page.getByRole('button', { name: 'Register repository' }).click();
	await expect(page.getByLabel('Access token')).toBeVisible();
});

test('a saved token the forge refuses shows the failure and stays saved', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await saveGitHubToken(page, 'ghp_revoked3333');
	await page.route('**/api/repos/discover', (route) =>
		route.fulfill({
			status: 502,
			json: { code: 'forge_error', message: 'GitHub refused the token.' },
		}),
	);

	await page.goto('/repos');
	await page.getByRole('button', { name: 'Register repository' }).click();

	await expect(page.getByText('GitHub refused the token.')).toBeVisible();
	await expect(page.getByLabel('Access token')).toBeVisible();

	await page.goto('/settings');
	await expect(page.getByRole('row', { name: /GitHub/ })).toContainText(
		'****3333',
	);
});

test('tokens the browser remembered before are removed, not read', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.evaluate(() => {
		localStorage.setItem('rememberedToken:github:', 'ghp_old0000');
		localStorage.setItem('lastForgeSelection', '{"forge":"github"}');
	});

	await page.goto('/repos');
	await page.getByRole('button', { name: 'Register repository' }).click();

	// The old copy is not offered as a saved token.
	await expect(page.getByLabel('Access token')).toBeVisible();
	expect(
		await page.evaluate(() => localStorage.getItem('rememberedToken:github:')),
	).toBeNull();
	expect(
		await page.evaluate(() => localStorage.getItem('lastForgeSelection')),
	).not.toBeNull();
});

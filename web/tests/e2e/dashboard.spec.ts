import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';

const a11yTags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

test('registers a passkey, sees the pipeline overview, logs out, then logs back in', async ({
	page,
	context,
}) => {
	// A CDP virtual authenticator stands in for a real passkey device --
	// there's no hardware authenticator in CI, and this is the same
	// mechanism Chrome DevTools' own WebAuthn panel uses.
	const cdp = await context.newCDPSession(page);
	await cdp.send('WebAuthn.enable');
	const { authenticatorId }: { authenticatorId: string } = await cdp.send(
		'WebAuthn.addVirtualAuthenticator',
		{
			options: {
				protocol: 'ctap2',
				transport: 'internal',
				hasResidentKey: true,
				hasUserVerification: true,
				isUserVerified: true,
				automaticPresenceSimulation: true,
			},
		},
	);

	await page.goto('/');
	await expect(page).toHaveURL(/\/login$/);

	const registerButton = page.getByRole('button', {
		name: 'Register your passkey',
	});
	await expect(registerButton).toBeVisible();

	const loginPageScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(loginPageScan.violations).toEqual([]);

	await registerButton.click();

	await expect(page).toHaveURL('/');
	await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible();
	await expect(page.getByText('No pipelines tracked yet.')).toBeVisible();

	const dashboardScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(dashboardScan.violations).toEqual([]);

	// The overview's own rendering logic (health badges, per-pipeline
	// cards) is exercised against a controlled response here -- seeding
	// real pipeline data would mean either a fake forge's webhook HMAC
	// signature (its secret is never exposed over HTTP) or reaching past
	// the API into the database directly, neither of which this journey
	// test's job is to do. What GET /api/pipelines returns for real data
	// is already covered by internal/httpserver/pipelines_test.go.
	await page.route('**/api/pipelines', (route) =>
		route.fulfill({
			json: [
				{
					id: 'healthy-1',
					repoId: 'repo-1',
					name: 'CI',
					healthStatus: 'healthy',
				},
				{
					id: 'unhealthy-1',
					repoId: 'repo-1',
					name: 'Deploy',
					healthStatus: 'unhealthy',
					triggeredSignals: ['failure_rate', 'flaky_step'],
				},
			],
		}),
	);
	await page.reload();

	await expect(page.getByRole('heading', { name: 'CI' })).toBeVisible();
	const ciItem = page.getByRole('listitem').filter({ hasText: 'CI' });
	await expect(ciItem).toContainText('healthy');

	await expect(page.getByRole('heading', { name: 'Deploy' })).toBeVisible();
	const deployItem = page.getByRole('listitem').filter({ hasText: 'Deploy' });
	await expect(deployItem).toContainText('unhealthy');
	await expect(deployItem).toContainText('elevated failure rate, flaky step');

	const overviewWithDataScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(overviewWithDataScan.violations).toEqual([]);

	await page.unroute('**/api/pipelines');
	await page.getByRole('button', { name: 'Log out' }).click();
	await expect(page).toHaveURL(/\/login$/);

	const loginButton = page.getByRole('button', { name: 'Log in with passkey' });
	await expect(loginButton).toBeVisible();
	await loginButton.click();

	await expect(page).toHaveURL('/');
	await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible();

	await cdp.send('WebAuthn.removeVirtualAuthenticator', { authenticatorId });
});

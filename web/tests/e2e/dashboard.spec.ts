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

	// Clicking through to a pipeline's detail page: the trend charts (task
	// 5.4) render duration and failure-rate history, and a regression has
	// to be visible in the chart itself, not just as a current aggregate
	// number -- so the fixture below deliberately has the last bucket
	// spike well above the earlier ones.
	const timestamps = Array.from({ length: 6 }, (_, i) =>
		new Date(Date.UTC(2026, 8, 1 + i)).toISOString(),
	);
	await page.route('**/api/pipelines/unhealthy-1', (route) =>
		route.fulfill({
			json: {
				id: 'unhealthy-1',
				repoId: 'repo-1',
				name: 'Deploy',
				healthStatus: 'unhealthy',
				triggeredSignals: ['duration_regression', 'failure_rate'],
				durationTrend: {
					timestamps,
					p50: [300, 305, 295, 300, 600, 650],
					p90: [420, 430, 410, 425, 900, 980],
				},
				failureRateTrend: {
					timestamps,
					rate: [0.05, 0.1, 0.05, 0.1, 0.3, 0.35],
				},
			},
		}),
	);
	// Step breakdown (tasks 5.5/5.6): one flaky step and one consistently
	// failing step, so the two get visibly distinct treatment, and the
	// flaky step's forgeUrl exercises the deep-link-to-forge requirement.
	await page.route('**/api/pipelines/unhealthy-1/steps', (route) =>
		route.fulfill({
			json: [
				{
					id: 'step-1',
					name: 'run tests',
					durationContributionSeconds: 420,
					queueSeconds: 30,
					execSeconds: 390,
					failureRate: 0,
					flaky: false,
				},
				{
					id: 'step-2',
					name: 'flaky integration test',
					durationContributionSeconds: 120,
					queueSeconds: 5,
					execSeconds: 115,
					failureRate: 0.3,
					flaky: true,
					forgeUrl: 'https://forge.example/owner/repo/actions/runs/1/job/2',
				},
				{
					id: 'step-3',
					name: 'deploy to prod',
					durationContributionSeconds: 60,
					queueSeconds: 2,
					execSeconds: 58,
					failureRate: 1,
					flaky: false,
				},
			],
		}),
	);
	await page.getByRole('link').filter({ hasText: 'Deploy' }).click();
	await expect(page).toHaveURL(/\/pipelines\/unhealthy-1$/);
	await expect(page.getByRole('heading', { name: 'Deploy' })).toBeVisible();
	await expect(
		page.getByText('duration regression, elevated failure rate'),
	).toBeVisible();
	// The legend confirms both series actually rendered, not just an
	// empty chart shell.
	await expect(page.getByText('p50', { exact: true })).toBeVisible();
	await expect(page.getByText('p90', { exact: true })).toBeVisible();

	// Slowest step first (5.5's duration ranking), and the flaky step is
	// visually distinguishable from the consistently-failing one.
	const stepRows = page.getByRole('row');
	await expect(stepRows.nth(1)).toContainText('run tests');
	const flakyRow = stepRows.filter({ hasText: 'flaky integration test' });
	await expect(flakyRow.getByText('flaky', { exact: true })).toBeVisible();
	const failingRow = stepRows.filter({ hasText: 'deploy to prod' });
	await expect(failingRow.getByText('failing', { exact: true })).toBeVisible();

	// Deep link to the originating forge (5.6).
	const forgeLink = page.getByRole('link', { name: 'View on forge' });
	await expect(forgeLink).toHaveAttribute(
		'href',
		'https://forge.example/owner/repo/actions/runs/1/job/2',
	);

	const detailScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(detailScan.violations).toEqual([]);

	await page.unroute('**/api/pipelines/unhealthy-1');
	await page.unroute('**/api/pipelines/unhealthy-1/steps');
	await page.getByRole('link', { name: 'All pipelines' }).click();
	await expect(page).toHaveURL('/');

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

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

	// Dark mode (cycles system -> light -> dark), persisted across a
	// reload, and re-scanned since contrast is theme-sensitive.
	const themeToggle = page.getByRole('button', { name: /Theme:/ });
	await themeToggle.click();
	await themeToggle.click();
	await expect(page.locator('html')).toHaveClass('dark');
	await page.reload();
	await expect(page.locator('html')).toHaveClass('dark');

	const darkModeScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(darkModeScan.violations).toEqual([]);

	await page.getByRole('button', { name: /Theme:/ }).click();
	await expect(page.locator('html')).not.toHaveClass('dark');

	// Repo management (register/list/untrack) is exercised against the
	// real POST/DELETE /api/repos endpoints, not mocked -- same reasoning
	// as the login/registration ceremony above: this is what those
	// handlers actually do, not a fixture standing in for them. The
	// forge-webhook step legitimately fails against a fake token, which
	// is itself part of what's being verified: a degraded repo still
	// registers and shows why.
	await page.getByRole('link', { name: 'Register a repository' }).click();
	await expect(page).toHaveURL('/repos');
	await expect(page.getByText('No repositories tracked yet.')).toBeVisible();

	await page.getByRole('button', { name: 'Register repository' }).click();
	await page.getByLabel('Access token').fill('ghp_faketoken1234');

	// The repo picker (discover-then-select) is exercised against a mocked
	// response -- the real GitHub/Forgejo API call it wraps is already
	// covered by the ListAccessibleRepos client tests.
	await page.route('**/api/repos/discover', (route) =>
		route.fulfill({
			json: ['alrayyes/demo-repo', 'alrayyes/dotfiles'],
		}),
	);
	await page.getByRole('button', { name: 'Find repositories' }).click();
	await expect(page.getByText('alrayyes/demo-repo')).toBeVisible();
	await page.unroute('**/api/repos/discover');

	await page.getByRole('button', { name: 'Register', exact: true }).click();

	const repoRow = page
		.getByRole('row')
		.filter({ hasText: 'alrayyes/demo-repo' });
	await expect(repoRow).toBeVisible();
	await expect(repoRow).toContainText('degraded');
	// The dialog's own closing animation leaves its (fading, near-invisible)
	// text in the DOM for a moment after submit -- axe scores that as a
	// real contrast failure if it catches the page mid-transition, so wait
	// for the dialog to actually finish closing first.
	await expect(page.getByText('Register a repository')).not.toBeVisible();

	const reposScan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
	expect(reposScan.violations).toEqual([]);

	await repoRow.getByRole('button', { name: 'Untrack' }).click();
	await expect(page.getByText('Untrack alrayyes/demo-repo?')).toBeVisible();
	await page
		.getByRole('button', { name: 'Untrack', exact: true })
		.last()
		.click();
	await expect(page.getByText('No repositories tracked yet.')).toBeVisible();

	// Token reuse (#72): the token used above is offered again rather than
	// having to be retyped, browser-local only.
	await page.getByRole('button', { name: 'Register repository' }).click();
	const useSavedToken = page.getByRole('button', {
		name: 'Use saved token (****1234)',
	});
	await expect(useSavedToken).toBeVisible();
	await useSavedToken.click();
	await expect(page.getByLabel('Access token')).toHaveValue(
		'ghp_faketoken1234',
	);
	await page.keyboard.press('Escape');

	await page.getByRole('link', { name: 'Pipelines' }).click();
	await expect(page).toHaveURL('/');

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

	// Usage view (task 5.7), reached from the pipeline detail page's repoId.
	await page.route('**/api/repos/repo-1/usage', (route) =>
		route.fulfill({
			json: [
				{ workflow: 'CI', runnerMinutes: 842.5 },
				{ workflow: 'Deploy', runnerMinutes: 120.2 },
			],
		}),
	);
	await page.getByRole('link', { name: 'Runner-minutes usage' }).click();
	await expect(page).toHaveURL('/repos/repo-1/usage');
	const usageRows = page.getByRole('row');
	await expect(usageRows.filter({ hasText: 'CI' })).toContainText('842.5');
	await expect(usageRows.filter({ hasText: 'Deploy' })).toContainText('120.2');

	const usageScan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
	expect(usageScan.violations).toEqual([]);

	await page.unroute('**/api/repos/repo-1/usage');
	await page.getByRole('link', { name: 'All pipelines' }).click();
	await expect(page).toHaveURL('/');

	// Release history (task: footer/releases layout fix) -- the page fetches
	// GitHub's own releases API directly, no backend proxy, so that's what
	// gets mocked here rather than an internal endpoint.
	await page.route(
		'https://api.github.com/repos/alrayyes/pipeline-analytics/releases',
		(route) =>
			route.fulfill({
				json: [
					{
						tag_name: 'v0.10.0',
						name: 'v0.10.0',
						html_url:
							'https://github.com/alrayyes/pipeline-analytics/releases/tag/v0.10.0',
						published_at: '2026-09-16T00:00:00Z',
						body: '### Features\n\n* add a dark mode toggle ([#61](https://github.com/alrayyes/pipeline-analytics/issues/61))',
					},
				],
			}),
	);
	await page.getByRole('link', { name: 'Release history' }).click();
	await expect(page).toHaveURL('/releases');
	await expect(page.getByRole('heading', { name: 'v0.10.0' })).toBeVisible();
	// The markdown body rendered as real HTML, not raw "### Features" text.
	await expect(page.getByRole('heading', { name: 'Features' })).toBeVisible();
	await expect(page.getByRole('link', { name: '#61' })).toBeVisible();

	const releasesScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(releasesScan.violations).toEqual([]);

	await page.unroute(
		'https://api.github.com/repos/alrayyes/pipeline-analytics/releases',
	);
	await page.getByRole('link', { name: 'Pipelines' }).click();
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

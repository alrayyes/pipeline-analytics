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
	await expect(page.getByText('No repositories registered yet')).toBeVisible();

	// Empty states by repo-registration status (#71): with zero repos
	// registered, the Pipelines/Repositories nav links are disabled (still
	// role="link" for assistive tech, matching Button.svelte's own disabled-
	// link convention -- distinguished here by aria-disabled/no href rather
	// than by role) and a single "Register a repository" CTA takes their
	// place. Scoped to <header> since the page itself also has a "Pipelines"
	// heading.
	const nav = page.locator('header');
	const disabledPipelinesLink = nav.getByText('Pipelines', { exact: true });
	const disabledReposLink = nav.getByText('Repositories', { exact: true });
	await expect(disabledPipelinesLink).toHaveAttribute('aria-disabled', 'true');
	await expect(disabledPipelinesLink).not.toHaveAttribute('href');
	await expect(disabledReposLink).toHaveAttribute('aria-disabled', 'true');
	await expect(disabledReposLink).not.toHaveAttribute('href');
	await expect(
		page.getByRole('link', { name: 'Register a repository' }),
	).toBeVisible();

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
	// covered by the ListAccessibleRepos client tests. Left routed for the
	// rest of this test, since the token-reuse retry below discovers again.
	await page.route('**/api/repos/discover', (route) =>
		route.fulfill({
			json: ['alrayyes/demo-repo', 'alrayyes/dotfiles'],
		}),
	);
	await page.getByRole('button', { name: 'Find repositories' }).click();

	// Multi-select (#103): check one discovered repo and follow it.
	const demoRepoCheckbox = page.getByRole('checkbox', {
		name: 'alrayyes/demo-repo',
	});
	await expect(demoRepoCheckbox).toBeVisible();
	await demoRepoCheckbox.check();
	await page.getByRole('button', { name: 'Follow 1 repository' }).click();

	const repoRow = page
		.getByRole('row')
		.filter({ hasText: 'alrayyes/demo-repo' });
	await expect(repoRow).toBeVisible();
	await expect(repoRow).toContainText('degraded');
	// The dialog's own closing animation leaves its (fading, near-invisible)
	// text in the DOM for a moment after submit -- axe scores that as a
	// real contrast failure if it catches the page mid-transition, so wait
	// for the dialog to actually finish closing first.
	await expect(
		page.getByText('Select repositories to follow'),
	).not.toBeVisible();

	const reposScan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
	expect(reposScan.violations).toEqual([]);

	// Empty states by repo-registration status (#71), the other half: a repo
	// is registered now, but nothing's been ingested yet -- the nav links go
	// back to normal, and the Pipelines page says so without telling the
	// user to register a repository again.
	await page.getByRole('link', { name: 'Pipelines', exact: true }).click();
	await expect(page).toHaveURL('/');
	await expect(
		page.getByRole('link', { name: 'Register a repository' }),
	).toHaveCount(0);
	await expect(page.getByText('No pipeline runs ingested yet')).toBeVisible();
	await page.getByRole('link', { name: 'Repositories', exact: true }).click();
	await expect(page).toHaveURL('/repos');

	await repoRow.getByRole('button', { name: 'Untrack' }).click();
	await expect(page.getByText('Untrack alrayyes/demo-repo?')).toBeVisible();
	await page
		.getByRole('button', { name: 'Untrack', exact: true })
		.last()
		.click();
	await expect(page.getByText('No repositories tracked yet.')).toBeVisible();

	// Token reuse (#72): the token used above is offered again rather than
	// having to be retyped, browser-local only. Registering for real here
	// (rather than abandoning the dialog) also puts a repo back so the rest
	// of this test's nav clicks -- disabled while zero repos are registered
	// (#71) -- keep working.
	await page.getByRole('button', { name: 'Register repository' }).click();
	const useSavedToken = page.getByRole('button', {
		name: 'Use saved token (****1234)',
	});
	await expect(useSavedToken).toBeVisible();
	// One click, not two (#117): filling the token and finding repositories
	// used to be two separate actions; using a saved token now goes straight
	// to the repo picker, replacing the token step's markup (including the
	// "Access token" field) with the picker's.
	await useSavedToken.click();
	await expect(
		page.getByRole('heading', { name: 'Select repositories to follow' }),
	).toBeVisible();

	// Registering more than one repo in a single batch (#103): one checked
	// from the discovered list, one added by name -- both land in the same
	// "Follow N repositories" submission.
	await page.getByRole('checkbox', { name: 'alrayyes/demo-repo' }).check();
	await page.getByLabel('Add another by name').fill('alrayyes/manual-repo');
	await page.getByRole('button', { name: 'Add', exact: true }).click();
	await expect(page.getByText('alrayyes/manual-repo')).toBeVisible();
	await page.getByRole('button', { name: 'Follow 2 repositories' }).click();

	await expect(repoRow).toBeVisible();
	await expect(
		page.getByRole('row').filter({ hasText: 'alrayyes/manual-repo' }),
	).toBeVisible();

	// Already-tracked repos (#117) don't belong in the picker a second
	// time -- demo-repo is tracked now, so re-discovering only offers
	// dotfiles.
	await page.getByRole('button', { name: 'Register repository' }).click();
	await page
		.getByRole('button', { name: 'Use saved token (****1234)' })
		.click();
	await expect(
		page.getByRole('checkbox', { name: 'alrayyes/dotfiles' }),
	).toBeVisible();
	await expect(
		page.getByRole('checkbox', { name: 'alrayyes/demo-repo' }),
	).toHaveCount(0);
	await page.keyboard.press('Escape');

	// Forge grouping and the persistent filter (#140): register one more
	// repo on Forgejo so the table actually has two forges to group, then
	// exercise the filter, then untrack it again -- the rest of this
	// journey doesn't expect a Forgejo repo hanging around.
	await page.getByRole('button', { name: 'Register repository' }).click();
	await page.getByLabel('Forge', { exact: true }).click();
	await page.getByRole('option', { name: 'Forgejo' }).click();
	await page.getByLabel('Access token').fill('forgejo_faketoken5678');
	await page
		.getByLabel('Forgejo instance URL')
		.fill('https://forgejo.example.com');
	await page.getByRole('button', { name: 'Find repositories' }).click();
	await page.getByLabel('Add another by name').fill('alrayyes/forgejo-repo');
	await page.getByRole('button', { name: 'Add', exact: true }).click();
	await page.getByRole('button', { name: 'Follow 1 repository' }).click();

	const githubHeading = page.getByRole('heading', {
		name: 'GitHub',
		exact: true,
	});
	const forgejoHeading = page.getByRole('heading', {
		name: 'Forgejo',
		exact: true,
	});
	await expect(githubHeading).toBeVisible();
	await expect(forgejoHeading).toBeVisible();
	const forgejoRepoRow = page
		.getByRole('row')
		.filter({ hasText: 'alrayyes/forgejo-repo' });
	await expect(forgejoRepoRow).toBeVisible();
	// Same dialog-closing-animation guard as the first registration above.
	await expect(
		page.getByText('Select repositories to follow'),
	).not.toBeVisible();

	const groupedReposScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(groupedReposScan.violations).toEqual([]);

	// The filter narrows the grouped table to one forge, applied server-side
	// (#140) -- not just hidden client-side.
	await page.getByRole('radio', { name: 'GitHub' }).click();
	await expect(githubHeading).toBeVisible();
	await expect(forgejoHeading).not.toBeVisible();
	await expect(forgejoRepoRow).not.toBeVisible();
	await expect(repoRow).toBeVisible();

	await page.getByRole('radio', { name: 'Forgejo' }).click();
	await expect(forgejoHeading).toBeVisible();
	await expect(githubHeading).not.toBeVisible();
	await expect(forgejoRepoRow).toBeVisible();
	await expect(repoRow).not.toBeVisible();

	// Persisted across navigation, shared with the Pipelines list page.
	await page.getByRole('link', { name: 'Pipelines', exact: true }).click();
	await expect(page).toHaveURL('/');
	await expect(page.getByRole('radio', { name: 'Forgejo' })).toHaveAttribute(
		'aria-checked',
		'true',
	);

	// Restore "All" before the rest of this journey, which expects an
	// unfiltered view, then go back and untrack the extra repo.
	await page.getByRole('radio', { name: 'All' }).click();
	await page.getByRole('link', { name: 'Repositories', exact: true }).click();
	await expect(page).toHaveURL('/repos');
	await forgejoRepoRow.getByRole('button', { name: 'Untrack' }).click();
	await expect(page.getByText('Untrack alrayyes/forgejo-repo?')).toBeVisible();
	await page
		.getByRole('button', { name: 'Untrack', exact: true })
		.last()
		.click();
	await expect(forgejoRepoRow).not.toBeVisible();

	await page.unroute('**/api/repos/discover');

	await page.getByRole('link', { name: 'Pipelines' }).click();
	await expect(page).toHaveURL('/');

	// The overview's own rendering logic (health badges, per-pipeline
	// cards) is exercised against a controlled response here -- seeding
	// real pipeline data would mean either a fake forge's webhook HMAC
	// signature (its secret is never exposed over HTTP) or reaching past
	// the API into the database directly, neither of which this journey
	// test's job is to do. What GET /api/pipelines returns for real data
	// is already covered by internal/httpserver/pipelines_test.go.
	// CI's last run is older than Deploy's -- distinct enough to tell the
	// "Name" and "Most recently run" sort orders apart below.
	const ciLastRunAt = new Date(
		Date.now() - 2 * 24 * 60 * 60 * 1000,
	).toISOString();
	const deployLastRunAt = new Date(Date.now() - 60 * 60 * 1000).toISOString();
	await page.route('**/api/pipelines', (route) =>
		route.fulfill({
			json: [
				{
					id: 'healthy-1',
					repoId: 'repo-1',
					name: 'CI',
					healthStatus: 'healthy',
					lastRunAt: ciLastRunAt,
				},
				{
					id: 'unhealthy-1',
					repoId: 'repo-1',
					name: 'Deploy',
					healthStatus: 'unhealthy',
					triggeredSignals: ['failure_rate', 'flaky_step'],
					lastRunAt: deployLastRunAt,
				},
			],
		}),
	);
	await page.reload();

	// Unhealthy-only by default (#101): the healthy "CI" pipeline is hidden
	// until the health filter is set to "All".
	const healthFilter = page.getByRole('radiogroup', {
		name: 'Filter by health status',
	});
	await expect(page.getByRole('heading', { name: 'CI' })).not.toBeVisible();
	await expect(page.getByRole('heading', { name: 'Deploy' })).toBeVisible();
	const deployItem = page.getByRole('listitem').filter({ hasText: 'Deploy' });
	await expect(deployItem).toContainText('unhealthy');
	await expect(deployItem).toContainText('elevated failure rate, flaky step');
	await expect(deployItem).toContainText('Last run');

	const filteredOverviewScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(filteredOverviewScan.violations).toEqual([]);

	await healthFilter.getByRole('radio', { name: 'All' }).click();
	await expect(page.getByRole('heading', { name: 'CI' })).toBeVisible();
	const ciItem = page.getByRole('listitem').filter({ hasText: 'CI' });
	await expect(ciItem).toContainText('healthy');

	// #149: "Name" (the default) sorts CI before Deploy alphabetically;
	// "Most recently run" reorders them the other way, since Deploy's
	// fixture lastRunAt is the more recent of the two.
	const pipelineNames = page.getByRole('heading', { level: 3 });
	await expect(pipelineNames).toHaveText(['CI', 'Deploy']);
	await page.getByLabel('Sort').click();
	await page.getByRole('option', { name: 'Most recently run' }).click();
	await expect(pipelineNames).toHaveText(['Deploy', 'CI']);
	// Same dialog-closing-animation guard as the repo registration dialog
	// above -- the dropdown's own fade-out otherwise reads as a real
	// contrast failure if axe catches it mid-transition.
	await expect(page.getByRole('listbox')).not.toBeVisible();

	const overviewWithDataScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(overviewWithDataScan.violations).toEqual([]);

	// The new success/destructive health badges are theme-sensitive colors
	// (#101) -- the earlier dark-mode scan ran before any pipeline data
	// existed, so it never actually rendered one. Set the theme directly
	// (rather than clicking the cycling toggle, whose position at this
	// point in the flow isn't guaranteed) and reload, then revert the same
	// way for the rest of the flow.
	await page.evaluate(() => localStorage.setItem('theme', 'dark'));
	await page.reload();
	await expect(page.locator('html')).toHaveClass('dark');
	// A reload is a fresh page load, so the unhealthy-only filter is back to
	// its default -- set it back to "All" to get both badge colors on screen.
	await healthFilter.getByRole('radio', { name: 'All' }).click();
	await expect(page.getByRole('heading', { name: 'CI' })).toBeVisible();
	const overviewDarkScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(overviewDarkScan.violations).toEqual([]);
	await page.evaluate(() => localStorage.setItem('theme', 'light'));
	await page.reload();
	await expect(page.locator('html')).not.toHaveClass('dark');

	// Grouped by repo (#102): a pipeline name alone ("CI") is ambiguous
	// across more than one tracked repo, so a second repo with its own "CI"
	// has to land in its own section, never merged with the first repo's.
	await page.route('**/api/repos', (route) =>
		route.fulfill({
			json: [
				{
					id: 'repo-1',
					forge: 'github',
					identifier: 'alrayyes/demo-repo',
					tokenMasked: '****1234',
					ingestionStatus: 'degraded',
				},
				{
					id: 'repo-2',
					forge: 'forgejo',
					identifier: 'alrayyes/other-repo',
					tokenMasked: '****5678',
					ingestionStatus: 'active',
				},
			],
		}),
	);
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
				{
					id: 'other-ci',
					repoId: 'repo-2',
					name: 'CI',
					healthStatus: 'unhealthy',
				},
			],
		}),
	);
	await page.reload();

	await expect(
		page.getByRole('heading', { name: 'alrayyes/demo-repo' }),
	).toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'alrayyes/other-repo' }),
	).toBeVisible();
	// Unhealthy-only by default: both repos' "Deploy"/"CI" show, but
	// repo-1's healthy "CI" stays hidden -- exactly one "CI" heading, not
	// two, confirming the two same-named pipelines didn't merge into one.
	await expect(
		page.getByRole('heading', { name: 'CI', exact: true }),
	).toHaveCount(1);
	const otherRepoSection = page.locator('section', {
		hasText: 'alrayyes/other-repo',
	});
	await expect(
		otherRepoSection.getByRole('heading', { name: 'CI' }),
	).toBeVisible();

	// #149: the repo filter narrows the grouped list to one repo, the same
	// way the forge filter already does.
	await page.getByLabel('Repo').click();
	await page.getByRole('option', { name: 'alrayyes/other-repo' }).click();
	await expect(
		page.getByRole('heading', { name: 'alrayyes/demo-repo' }),
	).not.toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'alrayyes/other-repo' }),
	).toBeVisible();
	await page.getByLabel('Repo').click();
	await page.getByRole('option', { name: 'All repos' }).click();
	// Same dialog-closing-animation guard as above -- the axe scan right
	// after this runs while the dropdown might still be fading out.
	await expect(page.getByRole('listbox')).not.toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'alrayyes/demo-repo' }),
	).toBeVisible();

	const groupingScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(groupingScan.violations).toEqual([]);

	// Restore the original two-pipeline fixture the rest of this journey
	// (including the /api/pipelines/unhealthy-1 detail mock below) expects.
	await page.unroute('**/api/repos');
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
	await expect(page.getByRole('heading', { name: 'Deploy' })).toBeVisible();

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
					name: 'Install dependencies and run the full integration test suite with coverage',
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
	const flakyRow = stepRows.filter({
		hasText:
			'Install dependencies and run the full integration test suite with coverage',
	});
	await expect(flakyRow.getByText('flaky', { exact: true })).toBeVisible();
	const failingRow = stepRows.filter({ hasText: 'deploy to prod' });
	await expect(failingRow.getByText('failing', { exact: true })).toBeVisible();
	// A step that's neither flaky nor failing gets its own "passing" badge
	// rather than an empty Status cell (#101).
	const passingRow = stepRows.filter({ hasText: 'run tests' });
	await expect(passingRow.getByText('passing', { exact: true })).toBeVisible();

	// #148: a realistically long step name (this fixture's "Install
	// dependencies..." step) shouldn't force the Steps table's own
	// overflow-x-auto container to actually scroll, on an ordinary desktop
	// viewport -- the long name is meant to truncate within its own column
	// instead of widening the whole table.
	const stepsOverflow = await page.evaluate(() => {
		const tableContainer = document.querySelector(
			'[data-slot="table-container"]',
		);

		return {
			scrollWidth: tableContainer?.scrollWidth ?? 0,
			clientWidth: tableContainer?.clientWidth ?? 0,
		};
	});
	expect(stepsOverflow.scrollWidth).toBeLessThanOrEqual(
		stepsOverflow.clientWidth,
	);

	// Deep link to the originating forge (5.6).
	const forgeLink = page.getByRole('link', { name: 'View on forge' });
	await expect(forgeLink).toHaveAttribute(
		'href',
		'https://forge.example/owner/repo/actions/runs/1/job/2',
	);

	// The whole row is clickable through to the forge, not just the small
	// link (#101) -- click near the row's start, away from the link itself.
	// Stubbing window.open rather than letting a real popup navigate to
	// forge.example, which isn't a real reachable host.
	await page.evaluate(() => {
		(window as unknown as { __openedUrls: string[] }).__openedUrls = [];
		window.open = (url) => {
			(window as unknown as { __openedUrls: string[] }).__openedUrls.push(
				String(url),
			);

			return null;
		};
	});
	await flakyRow.click({ position: { x: 10, y: 10 } });
	const openedUrls = await page.evaluate(
		() => (window as unknown as { __openedUrls: string[] }).__openedUrls,
	);
	expect(openedUrls).toEqual([
		'https://forge.example/owner/repo/actions/runs/1/job/2',
	]);

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

	// GitHub API rate-limit insights page -- the aggregation/grouping-by-
	// token logic is already covered by
	// internal/httpserver/insights_test.go; this exercises the page's own
	// rendering, including a token with no observed status yet.
	await page.route('**/api/insights/github-rate-limit', (route) =>
		route.fulfill({
			json: [
				{
					tokenMasked: '****1234',
					repos: ['alrayyes/demo-repo', 'alrayyes/other-repo'],
					status: {
						limit: 5000,
						remaining: 4900,
						used: 100,
						resetAt: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
					},
				},
				{
					tokenMasked: '****5678',
					repos: ['alrayyes/unseen-repo'],
				},
			],
		}),
	);
	await page.getByRole('link', { name: 'Insights' }).click();
	await expect(page).toHaveURL('/insights');
	await expect(page.getByRole('cell', { name: '****1234' })).toBeVisible();
	await expect(page.getByText('100 / 5000 (2%)')).toBeVisible();
	await expect(page.getByText(/in \d+ minutes?/)).toBeVisible();
	await expect(page.getByText('Not observed yet')).toBeVisible();

	const insightsScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(insightsScan.violations).toEqual([]);

	await page.unroute('**/api/insights/github-rate-limit');
	await page.getByRole('link', { name: 'Pipelines', exact: true }).click();
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
	// Long-form date (#73), matching forge-dashboard's own release history.
	await expect(page.getByText('September 16, 2026')).toBeVisible();

	const releasesScan = await new AxeBuilder({ page })
		.withTags(a11yTags)
		.analyze();
	expect(releasesScan.violations).toEqual([]);

	await page.unroute(
		'https://api.github.com/repos/alrayyes/pipeline-analytics/releases',
	);

	// Privacy & disclaimer (#78), reached from the footer everywhere else is.
	await page.getByRole('link', { name: 'Privacy & disclaimer' }).click();
	await expect(page).toHaveURL('/legal');
	await expect(
		page.getByRole('heading', { name: 'Privacy & disclaimer' }),
	).toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'Disclaimer', exact: true }),
	).toBeVisible();

	const legalScan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
	expect(legalScan.violations).toEqual([]);

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

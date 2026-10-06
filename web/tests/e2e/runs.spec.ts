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

const failedRun = {
	id: 'run-failed',
	pipelineId: 'p-deploy',
	pipelineName: 'main-deploy',
	repoId: 'r1',
	status: 'completed',
	conclusion: 'failure',
	outcome: 'failed',
	startedAt: '2026-10-02T12:00:00Z',
	durationSeconds: 402,
	branch: 'main',
	sha: 'c4d291a0e2b34f56a7c8d9e0f1a2b3c4d5e6f7a8',
	message: 'fix(stripe): webhook retry',
	actor: 'marcus-v',
	forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/1',
	actions: ['rerun'],
	steps: [
		{
			name: 'checkout',
			status: 'completed',
			conclusion: 'success',
			outcome: 'passed',
		},
		{
			name: 'build',
			status: 'completed',
			conclusion: 'success',
			outcome: 'passed',
		},
		{
			name: 'canary',
			status: 'completed',
			conclusion: 'timed_out',
			outcome: 'failed',
		},
		{ name: 'promote', status: 'queued', outcome: 'queued' },
	],
};

const runningRun = {
	id: 'run-running',
	pipelineId: 'p-preview',
	pipelineName: 'preview-build',
	repoId: 'r1',
	status: 'in_progress',
	outcome: 'running',
	startedAt: '2026-10-02T12:10:00Z',
	forgeUrl: 'https://github.com/alrayyes/web/actions/runs/2',
	actions: ['cancel'],
	steps: [
		{
			name: 'install',
			status: 'completed',
			conclusion: 'success',
			outcome: 'passed',
		},
		{ name: 'e2e-test', status: 'in_progress', outcome: 'running' },
		{ name: 'upload', status: 'queued', outcome: 'queued' },
	],
};

const passedRun = {
	id: 'run-passed',
	pipelineId: 'p-infra',
	pipelineName: 'terraform-plan',
	repoId: 'r1',
	status: 'completed',
	conclusion: 'success',
	outcome: 'passed',
	startedAt: '2026-10-02T11:00:00Z',
	durationSeconds: 115,
	forgeUrl: 'https://github.com/alrayyes/infra/actions/runs/3',
	actions: [],
	steps: [
		{
			name: 'init',
			status: 'completed',
			conclusion: 'success',
			outcome: 'passed',
		},
		{
			name: 'plan',
			status: 'completed',
			conclusion: 'success',
			outcome: 'passed',
		},
	],
};

// Answers GET /api/runs from `respond`, recording every requested URL so a
// test can assert on the status and offset the page asked for.
async function mockRuns(
	page: Page,
	respond: (url: URL) => { runs: unknown[]; hasMore?: boolean } | number,
): Promise<URL[]> {
	const requested: URL[] = [];

	await page.route(/\/api\/runs\?/, (route: Route) => {
		const url = new URL(route.request().url());
		requested.push(url);

		const body = respond(url);
		if (typeof body === 'number') return route.fulfill({ status: body });

		return route.fulfill({ json: { hasMore: false, ...body } });
	});

	return requested;
}

async function mockOneRepo(page: Page): Promise<void> {
	// The layout asks `/api/repos?limit=1`, so the glob needs the trailing `*`.
	await page.route('**/api/repos*', (route) =>
		route.fulfill({
			json: {
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			},
		}),
	);
}

test('lists runs with their stage progression, commit and forge link', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockRuns(page, () => ({ runs: [failedRun, runningRun, passedRun] }));

	await page.goto('/');
	await page.getByRole('link', { name: 'Runs', exact: true }).click();
	await expect(page).toHaveURL('/runs');
	await expect(
		page.getByRole('heading', { name: 'Runs', level: 1 }),
	).toBeVisible();

	const failed = page.getByRole('listitem').filter({ hasText: 'main-deploy' });
	await expect(failed.getByText('Failed', { exact: true })).toBeVisible();
	await expect(failed.getByText('Stage 3/4: canary')).toBeVisible();
	await expect(failed.getByText('c4d291a', { exact: true })).toBeVisible();
	await expect(failed.getByText('fix(stripe): webhook retry')).toBeVisible();
	await expect(failed.getByText('@marcus-v')).toBeVisible();
	await expect(failed.getByText('6.7m')).toBeVisible();
	await expect(failed.getByRole('link', { name: /view run/i })).toHaveAttribute(
		'href',
		'https://github.com/alrayyes/payments/actions/runs/1',
	);
	// Each stage is readable without seeing its colour.
	await expect(failed.getByText('build: Passed')).toBeAttached();
	await expect(failed.getByText('canary: Failed')).toBeAttached();
	await expect(failed.getByText('promote: Queued')).toBeAttached();

	const running = page
		.getByRole('listitem')
		.filter({ hasText: 'preview-build' });
	await expect(running.getByText('Running', { exact: true })).toBeVisible();
	await expect(running.getByText('Stage 2/3: e2e-test')).toBeVisible();

	const passed = page
		.getByRole('listitem')
		.filter({ hasText: 'terraform-plan' });
	await expect(passed.getByText('Passed', { exact: true })).toBeVisible();
	await expect(passed.getByText('2/2 passed')).toBeVisible();
	// Nothing known about the commit: no empty tag, no empty "@".
	await expect(passed.locator('code')).toHaveCount(0);
});

test('the status tabs ask the server for that status, from the first page', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	const requested = await mockRuns(page, (url) => ({
		runs:
			url.searchParams.get('status') === 'failed'
				? [failedRun]
				: [failedRun, passedRun],
	}));

	await page.goto('/runs');
	await expect(page.getByText('terraform-plan')).toBeVisible();

	await page.getByRole('radio', { name: 'Failed' }).click();

	await expect(page.getByText('terraform-plan')).toHaveCount(0);
	const last = requested[requested.length - 1];
	expect(last.searchParams.get('status')).toBe('failed');
	expect(last.searchParams.get('offset')).toBe('0');
	// "all" is the server default and stays off the URL.
	expect(requested[0].searchParams.has('status')).toBe(false);
});

test('paging asks for the next page of 20', async ({ page, context }) => {
	await signIn(page, context);
	const requested = await mockRuns(page, () => ({
		runs: [failedRun],
		hasMore: true,
	}));

	await page.goto('/runs');
	await expect(page.getByText('main-deploy')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Previous' })).toBeDisabled();

	await page.getByRole('button', { name: 'Next' }).click();

	await expect
		.poll(() => requested[requested.length - 1].searchParams.get('offset'))
		.toBe('20');
	expect(requested[0].searchParams.get('limit')).toBe('20');
});

test('refreshes every 15 seconds while a run is going, then stops', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.clock.install();

	let stillRunning = true;
	const requested = await mockRuns(page, () => ({
		runs: stillRunning ? [runningRun] : [passedRun],
	}));

	await page.goto('/runs');
	await expect(page.getByText('preview-build')).toBeVisible();
	const afterLoad = requested.length;

	await page.clock.fastForward(15_000);
	await expect.poll(() => requested.length).toBeGreaterThan(afterLoad);

	// The run finished: the next refresh shows it, and nothing follows.
	stillRunning = false;
	await page.clock.fastForward(15_000);
	await expect(page.getByText('terraform-plan')).toBeVisible();
	const settled = requested.length;

	await page.clock.fastForward(60_000);
	expect(requested.length).toBe(settled);
});

test('says so when nothing matches, and when the list cannot load', async ({
	page,
	context,
}) => {
	await signIn(page, context);

	let status: number | null = null;
	await mockRuns(page, () => status ?? { runs: [] });

	await page.goto('/runs');
	await expect(page.getByText('No runs match')).toBeVisible();

	status = 500;
	await page.getByRole('radio', { name: 'Failed' }).click();
	await expect(page.getByRole('alert')).toContainText("Couldn't load runs");
});

for (const scheme of ['light', 'dark'] as const) {
	test(`fits a phone and passes axe in ${scheme} mode`, async ({
		page,
		context,
	}) => {
		await signIn(page, context);
		await page.setViewportSize({ width: 390, height: 844 });
		await page.emulateMedia({ colorScheme: scheme });
		await mockRuns(page, () => ({
			runs: [failedRun, runningRun, passedRun],
			hasMore: true,
		}));

		await page.goto('/runs');
		await expect(page.getByText('main-deploy')).toBeVisible();

		const overflow = await page.evaluate(
			() => document.documentElement.scrollWidth - window.innerWidth,
		);
		expect(overflow).toBeLessThanOrEqual(0);

		const scan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
		expect(scan.violations).toEqual([]);
	});
}

test('re-runs a run the server offers it for, and says why when refused', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await mockRuns(page, () => ({ runs: [failedRun, runningRun, passedRun] }));

	const answers = [
		{ status: 403, json: { code: 'forbidden', message: 'read-only' } },
		{ status: 202 },
	];
	const posted: string[] = [];
	await page.route('**/api/runs/*/rerun', (route) => {
		posted.push(new URL(route.request().url()).pathname);

		return route.fulfill(answers.shift() ?? { status: 202 });
	});

	await page.goto('/runs');

	// Only the run the server lists `rerun` for gets the button.
	await expect(page.getByRole('button', { name: 'Re-run' })).toHaveCount(1);
	const failed = page.getByRole('listitem').filter({ hasText: 'main-deploy' });

	await failed.getByRole('button', { name: 'Re-run' }).click();
	await expect(failed.getByRole('alert')).toContainText('Actions write');

	await failed.getByRole('button', { name: 'Re-run' }).click();
	await expect(failed.getByRole('status')).toContainText('Re-run requested');
	expect(posted).toEqual([
		'/api/runs/run-failed/rerun',
		'/api/runs/run-failed/rerun',
	]);

	const scan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
	expect(scan.violations).toEqual([]);
});

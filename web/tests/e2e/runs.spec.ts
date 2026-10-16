import type { BrowserContext, Page, Route } from '@playwright/test';
import { expectFitsAndPassesAxeInBothSchemes, scanSettled } from './axe.js';
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

test('fits a phone and passes axe in light and dark mode', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.setViewportSize({ width: 390, height: 844 });
	await mockRuns(page, () => ({
		runs: [failedRun, runningRun, passedRun],
		hasMore: true,
	}));

	await page.goto('/runs');
	await expect(page.getByText('main-deploy')).toBeVisible();

	await expectFitsAndPassesAxeInBothSchemes(page, async () => {
		await expect(page.getByText('main-deploy')).toBeVisible();
	});
});

// The step log viewer: the last lines of a failed step's job, ANSI colours
// drawn, and never anything the log says treated as markup.
function runWithFailedStep() {
	return {
		runId: 'run-1',
		startedAt: '2026-10-02T12:00:00Z',
		steps: [
			{
				jobId: 'job-1',
				name: 'checkout',
				status: 'completed',
				conclusion: 'success',
				outcome: 'passed',
			},
			{
				jobId: 'job-2',
				name: 'canary',
				status: 'completed',
				conclusion: 'failure',
				outcome: 'failed',
				forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/1/job/2',
			},
		],
	};
}

test('a failed step shows its log, coloured, and offers none for a passing one', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.route('**/api/runs/run-1/steps', (route) =>
		route.fulfill({ json: runWithFailedStep() }),
	);
	let asked = '';
	await page.route('**/api/runs/run-1/jobs/job-2/log*', (route) => {
		asked = new URL(route.request().url()).search;

		return route.fulfill({
			json: {
				available: true,
				lines: ['\u001b[31merror:\u001b[0m canary timed out', 'exit code 1'],
				truncated: true,
				forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/1/job/2',
			},
		});
	});

	await page.goto('/runs/run-1');
	await expect(page.getByRole('button', { name: 'Show log' })).toHaveCount(1);

	await page.getByRole('button', { name: 'Show log' }).click();
	const log = page.getByLabel('Step log');
	await expect(log).toContainText('error: canary timed out');
	expect(asked).toBe('?lines=200');

	// The escape code is drawn as a colour, not left in the text.
	await expect(log).not.toContainText('[31m');
	const colour = await log
		.getByText('error:')
		.evaluate((el) => getComputedStyle(el).color);
	expect(colour).toBe('rgb(255, 123, 123)');

	await expect(log.locator('..').getByText('Last 200 lines.')).toBeVisible();

	const scan = await scanSettled(page);
	expect(scan.violations).toEqual([]);
});

test('markup in a log is shown as text and never runs', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.route('**/api/runs/run-1/steps', (route) =>
		route.fulfill({ json: runWithFailedStep() }),
	);
	const payload =
		'<script>window.__pwned = 1</script><img src=x onerror="window.__pwned = 2">';
	await page.route('**/api/runs/run-1/jobs/job-2/log*', (route) =>
		route.fulfill({
			json: {
				available: true,
				lines: [payload],
				truncated: false,
				forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/1/job/2',
			},
		}),
	);

	await page.goto('/runs/run-1');
	await page.getByRole('button', { name: 'Show log' }).click();
	const log = page.getByLabel('Step log');

	await expect(log).toContainText(payload);
	await expect(log.locator('script, img')).toHaveCount(0);
	expect(
		await page.evaluate(() => (window as { __pwned?: number }).__pwned),
	).toBe(undefined);
});

test('a log the forge cannot give says why and still links to the forge', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.route('**/api/runs/run-1/steps', (route) =>
		route.fulfill({ json: runWithFailedStep() }),
	);
	await page.route('**/api/runs/run-1/jobs/job-2/log*', (route) =>
		route.fulfill({
			json: {
				available: false,
				reason: 'unsupported',
				lines: [],
				truncated: false,
				forgeUrl: 'https://forge.example/alrayyes/payments/actions/runs/1',
			},
		}),
	);

	await page.goto('/runs/run-1');
	await page.getByRole('button', { name: 'Show log' }).click();

	await expect(page.getByText('no log API')).toBeVisible();
	await expect(page.getByLabel('Step log')).toHaveCount(0);
	await expect(
		page.getByRole('link', { name: 'View on forge' }).last(),
	).toHaveAttribute(
		'href',
		'https://forge.example/alrayyes/payments/actions/runs/1',
	);

	const scan = await scanSettled(page);
	expect(scan.violations).toEqual([]);
});

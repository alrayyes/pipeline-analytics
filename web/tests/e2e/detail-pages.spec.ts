import { signIn } from './auth.js';
import { expectNoViolations } from './axe.js';
import { expect, test } from './fixtures.js';

// The detail pages are reached from a list, so nothing else visits them: one
// journey each, answered from fixtures, with the axe scan once the data is on
// screen (a scan of "Loading…" would pass for the wrong reason).

const trend = {
	timestamps: ['2026-10-01T00:00:00Z', '2026-10-02T00:00:00Z'],
	p50: [100, 120],
	p90: [200, 240],
	rate: [0.1, 0.2],
};

test('a pipeline passes axe', async ({ page, context }) => {
	await signIn(page, context);
	await page.route('**/api/pipelines/p1', (route) =>
		route.fulfill({
			json: {
				id: 'p1',
				repoId: 'r1',
				name: 'CI',
				healthStatus: 'unhealthy',
				triggeredSignals: ['failure_rate'],
				durationTrend: trend,
				failureRateTrend: trend,
			},
		}),
	);
	await page.route('**/api/pipelines/p1/steps*', (route) =>
		route.fulfill({
			json: [
				{
					id: 's1',
					name: 'build',
					durationContributionSeconds: 90,
					queueSeconds: 5,
					execSeconds: 85,
					failureRate: 0.2,
					failureCount: 2,
					flaky: true,
					forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/1',
				},
			],
		}),
	);

	await page.goto('/pipelines/p1');
	await expect(page.getByRole('heading', { name: 'CI' })).toBeVisible();
	await expect(page.getByRole('cell', { name: 'build' })).toBeVisible();

	await expectNoViolations(page);
});

test('a quarantined flaky step is labelled on the pipeline page, on a phone too', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.setViewportSize({ width: 390, height: 800 });
	await page.route('**/api/pipelines/p1', (route) =>
		route.fulfill({
			json: {
				id: 'p1',
				repoId: 'r1',
				name: 'CI',
				healthStatus: 'healthy',
				durationTrend: trend,
				failureRateTrend: trend,
			},
		}),
	);
	const step = {
		durationContributionSeconds: 90,
		queueSeconds: 5,
		execSeconds: 85,
		failureRate: 0.2,
		failureCount: 2,
		flaky: true,
	};
	await page.route('**/api/pipelines/p1/steps*', (route) =>
		route.fulfill({
			json: [
				{
					...step,
					id: 's1',
					name: 'browser tests',
					quarantined: true,
					quarantine: {
						note: 'Waiting on the vendor fix',
						quarantinedAt: '2026-10-02T00:00:00Z',
						expiresAt: '2026-11-01T00:00:00Z',
					},
				},
				{ ...step, id: 's2', name: 'unit tests', quarantined: false },
			],
		}),
	);

	await page.goto('/pipelines/p1');

	const quarantined = page.getByRole('row', { name: /browser tests/ });
	await expect(quarantined).toContainText('flaky');
	await expect(quarantined).toContainText('Quarantined');
	await expect(page.getByRole('row', { name: /unit tests/ })).not.toContainText(
		'Quarantined',
	);

	const overflow = await page.evaluate(
		() => document.documentElement.scrollWidth - window.innerWidth,
	);
	expect(overflow).toBeLessThanOrEqual(0);
	await expectNoViolations(page);
});

test('a run passes axe', async ({ page, context }) => {
	await signIn(page, context);
	await page.route('**/api/runs/run-1/steps', (route) =>
		route.fulfill({
			json: {
				runId: 'run-1',
				startedAt: '2026-10-02T12:00:00Z',
				steps: [
					{
						name: 'checkout',
						status: 'completed',
						conclusion: 'success',
						forgeUrl: 'https://github.com/alrayyes/payments/actions/runs/1',
					},
					{ name: 'canary', status: 'completed', conclusion: 'failure' },
				],
			},
		}),
	);

	await page.goto('/runs/run-1');
	await expect(page.getByText('canary')).toBeVisible();

	await expectNoViolations(page);
});

test('a repository’s runner usage passes axe', async ({ page, context }) => {
	await signIn(page, context);
	await page.route('**/api/repos/r1/usage', (route) =>
		route.fulfill({
			json: [
				{ workflow: 'CI', runnerMinutes: 120 },
				{ workflow: 'Release', runnerMinutes: 14 },
			],
		}),
	);

	await page.goto('/repos/r1/usage');
	await expect(page.getByText('Release')).toBeVisible();

	await expectNoViolations(page);
});

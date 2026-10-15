import '../../test/dom.js';
import {
	afterAll,
	afterEach,
	beforeEach,
	describe,
	expect,
	test,
} from 'bun:test';
import { mount, unmount } from 'svelte';
import { initTelemetryWindow } from '#lib/telemetryWindow.svelte.js';
import { removeDom } from '../../test/dom.js';
import { fakeNetwork, json, radio, settle } from '../../test/page.js';
import Page from './+page.svelte';

// The Overview page against a fake network: what its window toggle asks the
// server for, and what it says when there is nothing to show, no data for a
// figure, or the request fails. The figures themselves come from the server;
// this is the page's own wiring.

const insights = {
	window: '7d',
	totalRuns: 658,
	failedRuns: 142,
	passRate: 0.784,
	passRateDelta: -4.2,
	flakyStepRatio: 0.068,
	mttrSeconds: 2280,
	stageDistribution: [],
	categoryBreakdown: [],
	topFailingPipelines: [],
	failureGroups: [],
};

const latestFailure = {
	id: 'run-9',
	pipelineId: 'p-deploy',
	pipelineName: 'staging-deploy',
	repoId: 'r1',
	status: 'completed',
	conclusion: 'failure',
	outcome: 'failed',
	startedAt: '2026-10-02T12:00:00Z',
	sha: 'c4d291a0e2b34f56a7c8d9e0f1a2b3c4d5e6f7a8',
	message: 'fix(stripe): webhook retry',
	steps: [],
};

let network: ReturnType<typeof fakeNetwork>;
let respondInsights: (url: URL) => Response;
let failedRuns: unknown[];
let app: ReturnType<typeof mount> | undefined;

function insightRequests(): URL[] {
	return network.requested.filter(
		(url) => url.pathname === '/api/insights/failures',
	);
}

beforeEach(() => {
	localStorage.clear();
	initTelemetryWindow();
	respondInsights = () => json(insights);
	failedRuns = [];

	network = fakeNetwork((url) => {
		if (url.pathname === '/api/insights/failures') return respondInsights(url);
		if (url.pathname === '/api/runs') {
			return json({ runs: failedRuns, hasMore: false });
		}
		if (url.pathname === '/api/branches') return json({ branches: [] });
	});
});

afterEach(() => {
	if (app) unmount(app);
	app = undefined;
	document.body.innerHTML = '';
});

afterAll(async () => {
	network.restore();
	await removeDom();
});

async function open(hasRepos = true): Promise<void> {
	app = mount(Page, { target: document.body, props: { data: { hasRepos } } });
	await settle();
}

describe('Overview page', () => {
	test('with no repository registered it says how to start and asks for nothing', async () => {
		await open(false);

		expect(document.body.textContent).toContain(
			'No repositories registered yet',
		);
		expect(network.requested).toEqual([]);
	});

	test('the window toggle asks the server for that window', async () => {
		await open();

		radio('24 hours').click();
		await settle();

		expect(insightRequests().at(-1)?.searchParams.get('window')).toBe('24h');
		expect(radio('24 hours').getAttribute('aria-checked')).toBe('true');
	});

	test('a figure the server has no data for says so instead of showing zero', async () => {
		respondInsights = () =>
			json({
				...insights,
				passRate: undefined,
				passRateDelta: undefined,
				mttrSeconds: undefined,
				totalRuns: 3,
				failedRuns: 0,
			});

		await open();

		expect(document.body.textContent).toContain('No data');
		expect(document.body.textContent).toContain('No recoveries');
		expect(document.querySelector('#latest-failure-heading')).toBeNull();
	});

	test('shows the latest failure when there is one', async () => {
		failedRuns = [latestFailure];

		await open();

		// The same element the no-data test asserts is absent, so that assertion
		// looks at something that really does appear.
		expect(document.querySelector('#latest-failure-heading')).not.toBeNull();
		expect(document.body.textContent).toContain('staging-deploy');
	});

	test('says so when the figures cannot load', async () => {
		respondInsights = () => json({}, 500);

		await open();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			"Couldn't load the overview",
		);
	});
});

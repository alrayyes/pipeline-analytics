import '../../../test/dom.js';
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
import { removeDom } from '../../../test/dom.js';
import { fakeNetwork, json, radio, settle } from '../../../test/page.js';
import Page from './+page.svelte';

// The Failures page against a fake network: what the window toggle asks the
// server for, and what the page says when there is nothing or the request
// fails. The rules behind those answers are the server's; this is the page's
// own wiring, which only a browser test could reach before.

const insights = {
	window: '7d',
	totalRuns: 10,
	failedRuns: 2,
	passRate: 0.8,
	categoryBreakdown: [],
	failureGroups: [],
};

let network: ReturnType<typeof fakeNetwork>;
let respond: (url: URL) => Response;
let app: ReturnType<typeof mount> | undefined;

function insightRequests(): URL[] {
	return network.requested.filter(
		(url) => url.pathname === '/api/insights/failures',
	);
}

beforeEach(() => {
	localStorage.clear();
	// The window is shared state across the app; start each test with none saved.
	initTelemetryWindow();
	respond = () => json(insights);

	network = fakeNetwork((url) => {
		if (url.pathname === '/api/insights/failures') return respond(url);
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

async function open(): Promise<void> {
	app = mount(Page, { target: document.body });
	await settle();
}

describe('Failures page', () => {
	test('the window toggle asks the server for that window', async () => {
		await open();

		radio('30 days').click();
		await settle();

		const last = insightRequests().at(-1);
		expect(last?.searchParams.get('window')).toBe('30d');
		expect(radio('30 days').getAttribute('aria-checked')).toBe('true');
	});

	test('with no saved window it asks for none, and shows the one the server used', async () => {
		respond = () => json({ ...insights, window: '24h' });

		await open();

		expect(insightRequests()[0]?.searchParams.has('window')).toBe(false);
		expect(radio('24 hours').getAttribute('aria-checked')).toBe('true');
	});

	test('says so when nothing failed', async () => {
		await open();

		expect(document.body.textContent).toContain('No failures in this window.');
		expect(document.querySelector('#failure-categories-heading')).toBeNull();
	});

	test('shows the category breakdown when something failed', async () => {
		respond = () =>
			json({
				...insights,
				failedRuns: 2,
				categoryBreakdown: [
					{ category: 'code_tests', occurrences: 2, share: 1 },
				],
				failureGroups: [
					{
						step: 'Run unit tests',
						category: 'code_tests',
						occurrences: 2,
						pipelines: [{ pipelineId: 'p1', pipelineName: 'api-gateway' }],
					},
				],
			});

		await open();

		// The same element the empty state asserts is absent, so that assertion
		// is looking at something that really does appear.
		expect(
			document.querySelector('#failure-categories-heading'),
		).not.toBeNull();
	});

	test('says so when the figures cannot load', async () => {
		await open();

		respond = () => json({}, 500);
		radio('24 hours').click();
		await settle();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			"Couldn't load failures",
		);
	});
});

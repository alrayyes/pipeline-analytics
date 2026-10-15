import '../../../../../test/dom.js';
import {
	afterAll,
	afterEach,
	beforeEach,
	describe,
	expect,
	mock,
	test,
} from 'bun:test';
import { mount, unmount } from 'svelte';
import { removeDom } from '../../../../../test/dom.js';
import { fakeNetwork, json, settle } from '../../../../../test/page.js';

// The flaky-runs page against a fake network: what it asks for, and what it says
// for runs found, none, a missing pipeline, a failed load or an unreachable
// server. The page and the list_pipeline_flaky_runs WebMCP tool share one fetch
// function, so this keeps the page's messages as they were.

mock.module('$app/state', () => ({
	page: {
		params: { id: 'p 1' },
		url: new URL(
			'http://localhost/pipelines/p%201/flaky-runs?step=Run+unit+tests',
		),
	},
}));
mock.module('$app/navigation', () => ({ goto: () => Promise.resolve() }));

const { default: Page } = await import('./+page.svelte');

let network: ReturnType<typeof fakeNetwork>;
let answer: () => Response;
let app: ReturnType<typeof mount> | undefined;

beforeEach(() => {
	answer = () => json([{ runId: 'run-a', startedAt: '2026-10-02T12:00:00Z' }]);
	network = fakeNetwork((url) => {
		if (url.pathname.endsWith('/flaky-runs')) return answer();
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

function alertText(): string | undefined {
	return document.querySelector('[role="alert"]')?.textContent ?? undefined;
}

describe('Flaky runs page', () => {
	test('asks for the pipeline and step, escaped, and lists the runs', async () => {
		await open();

		const asked = network.requested[0];
		expect(asked?.pathname).toBe('/api/pipelines/p%201/flaky-runs');
		expect(asked?.searchParams.get('step')).toBe('Run unit tests');
		expect(document.querySelector('a[href^="/runs/run-a?"]')).not.toBeNull();
		expect(alertText()).toBeUndefined();
	});

	test('says so when no run flaked', async () => {
		answer = () => json([]);

		await open();

		expect(document.body.textContent).toContain(
			'No flaky runs in this window.',
		);
	});

	test('says so when the pipeline does not exist', async () => {
		answer = () => json({ code: 'not_found' }, 404);

		await open();

		expect(alertText()).toContain('Pipeline not found.');
	});

	test('says so when the runs cannot be loaded', async () => {
		answer = () => json({}, 500);

		await open();

		expect(alertText()).toContain('Could not load flaky runs.');
	});

	test('says so when the server cannot be reached', async () => {
		answer = () => {
			throw new TypeError('network down');
		};

		await open();

		expect(alertText()).toContain('Could not reach the server.');
	});
});

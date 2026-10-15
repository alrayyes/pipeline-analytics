import '../../../../test/dom.js';
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
import { removeDom } from '../../../../test/dom.js';
import { fakeNetwork, json, settle } from '../../../../test/page.js';

// The run page against a fake network: what it asks for, and what it says for a
// run that is found, missing, failed to load or unreachable. The page and the
// get_run_steps WebMCP tool share one fetch function, so this is what keeps the
// page's messages as they were.

mock.module('$app/state', () => ({
	page: {
		params: { id: 'run 1' },
		url: new URL('http://localhost/runs/run%201'),
	},
}));

const { default: Page } = await import('./+page.svelte');

const detail = {
	runId: 'run 1',
	startedAt: '2026-10-02T12:00:00Z',
	steps: [
		{
			name: 'checkout',
			status: 'completed',
			conclusion: 'success',
			outcome: 'passed',
		},
		{
			name: 'canary',
			status: 'completed',
			conclusion: 'failure',
			outcome: 'failed',
		},
	],
};

let network: ReturnType<typeof fakeNetwork>;
let answer: () => Response | Promise<Response>;
let app: ReturnType<typeof mount> | undefined;

beforeEach(() => {
	answer = () => json(detail);
	network = fakeNetwork((url) => {
		if (url.pathname.endsWith('/steps')) return answer() as Response;
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

describe('Run page', () => {
	test('asks for the run by its id, escaped, and lists its steps', async () => {
		await open();

		expect(network.requested[0]?.pathname).toBe('/api/runs/run%201/steps');
		expect(document.body.textContent).toContain('canary');
		expect(document.querySelector('[role="alert"]')).toBeNull();
	});

	test('says so when the run does not exist', async () => {
		answer = () => json({ code: 'not_found' }, 404);

		await open();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			'Run not found.',
		);
	});

	test('says so when the run cannot be loaded', async () => {
		answer = () => json({}, 500);

		await open();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			'Could not load run.',
		);
	});

	test('says so when the server cannot be reached', async () => {
		answer = () => {
			throw new TypeError('network down');
		};

		await open();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			'Could not reach the server.',
		);
	});
});

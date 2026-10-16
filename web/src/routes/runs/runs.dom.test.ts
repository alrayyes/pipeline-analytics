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
import { resetRunsFilters } from '#lib/runsFilters.svelte.js';
import { removeDom } from '../../../test/dom.js';
import { fakeNetwork, json, radio, settle } from '../../../test/page.js';
import Page from './+page.svelte';

// The Runs page against a fake network: what its status tabs and pager ask the
// server for, when it refreshes itself, what it says when nothing matches or the
// list cannot load, and what the re-run and cancel buttons do with a refusal. The
// list itself and the stage progression are the server's and the card's own tests.

const failedRun = {
	id: 'run-failed',
	pipelineId: 'p-deploy',
	pipelineName: 'main-deploy',
	repoId: 'r1',
	status: 'completed',
	conclusion: 'failure',
	outcome: 'failed',
	startedAt: '2026-10-02T12:00:00Z',
	actions: ['rerun'],
	steps: [{ name: 'canary', status: 'completed', outcome: 'failed' }],
};

const runningRun = {
	id: 'run-running',
	pipelineId: 'p-preview',
	pipelineName: 'preview-build',
	repoId: 'r1',
	status: 'in_progress',
	outcome: 'running',
	startedAt: '2026-10-02T12:10:00Z',
	actions: ['cancel'],
	steps: [{ name: 'e2e-test', status: 'in_progress', outcome: 'running' }],
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
	actions: [],
	steps: [{ name: 'plan', status: 'completed', outcome: 'passed' }],
};

let network: ReturnType<typeof fakeNetwork>;
let respond: (url: URL) => Response;
let action: (path: string) => Response;
let posted: string[];
let app: ReturnType<typeof mount> | undefined;

// The refreshes the page schedules, held back so a test fires them by hand.
const realSetTimeout = globalThis.setTimeout;
let scheduled: { fn: () => void; ms: number }[];

function runsRequests(): URL[] {
	return network.requested.filter((url) => url.pathname === '/api/runs');
}

beforeEach(() => {
	resetRunsFilters();
	posted = [];
	respond = () =>
		json({ runs: [failedRun, runningRun, passedRun], hasMore: false });
	action = () => json({}, 202);
	scheduled = [];

	globalThis.setTimeout = ((fn: () => void, ms?: number) => {
		if (ms === 15_000) {
			scheduled.push({ fn, ms });

			return 0;
		}

		return realSetTimeout(fn, ms);
	}) as unknown as typeof setTimeout;

	network = fakeNetwork((url, init) => {
		if (url.pathname === '/api/runs') return respond(url);
		if (url.pathname === '/api/branches') return json({ branches: [] });
		if (init?.method === 'POST') {
			posted.push(url.pathname);

			return action(url.pathname);
		}
	});
});

afterEach(() => {
	if (app) unmount(app);
	app = undefined;
	document.body.innerHTML = '';
	globalThis.setTimeout = realSetTimeout;
});

afterAll(async () => {
	network.restore();
	await removeDom();
});

async function open(): Promise<void> {
	app = mount(Page, { target: document.body });
	await settle();
}

function button(name: string, within: ParentNode = document): HTMLElement {
	const found = [...within.querySelectorAll<HTMLElement>('button')].find(
		(el) => el.textContent?.trim() === name,
	);
	if (!found) throw new Error(`no button named ${name}`);

	return found;
}

function card(name: string): HTMLElement {
	const found = [...document.querySelectorAll<HTMLElement>('li')].find((el) =>
		el.textContent?.includes(name),
	);
	if (!found) throw new Error(`no run named ${name}`);

	return found;
}

describe('Runs page', () => {
	test('the status tabs ask the server for that status, from the first page', async () => {
		respond = (url) =>
			json({
				runs:
					url.searchParams.get('status') === 'failed'
						? [failedRun]
						: [failedRun, passedRun],
				hasMore: false,
			});
		await open();
		expect(document.body.textContent).toContain('terraform-plan');

		radio('Failed').click();
		await settle();

		expect(document.body.textContent).not.toContain('terraform-plan');
		expect(runsRequests().at(-1)?.searchParams.get('status')).toBe('failed');
		expect(runsRequests().at(-1)?.searchParams.get('offset')).toBe('0');
		// "all" is the server default and stays off the URL.
		expect(runsRequests()[0]?.searchParams.has('status')).toBe(false);
	});

	test('paging asks for the next page of 20', async () => {
		respond = () => json({ runs: [failedRun], hasMore: true });
		await open();
		expect(button('Previous').hasAttribute('disabled')).toBe(true);

		button('Next').click();
		await settle();

		expect(runsRequests()[0]?.searchParams.get('limit')).toBe('20');
		expect(runsRequests().at(-1)?.searchParams.get('offset')).toBe('20');
	});

	test('refreshes every 15 seconds while a run is going, then stops', async () => {
		let stillRunning = true;
		respond = () =>
			json({ runs: stillRunning ? [runningRun] : [passedRun], hasMore: false });
		await open();
		const afterLoad = runsRequests().length;
		expect(scheduled).toHaveLength(1);

		scheduled.shift()?.fn();
		await settle();
		expect(runsRequests().length).toBeGreaterThan(afterLoad);

		// The run finished: the next refresh shows it, and nothing follows.
		stillRunning = false;
		expect(scheduled).toHaveLength(1);
		scheduled.shift()?.fn();
		await settle();

		expect(document.body.textContent).toContain('terraform-plan');
		expect(scheduled).toHaveLength(0);
	});

	test('says so when nothing matches, and when the list cannot load', async () => {
		respond = () => json({ runs: [], hasMore: false });
		await open();
		expect(document.body.textContent).toContain('No runs match');

		respond = () => json({}, 500);
		radio('Failed').click();
		await settle();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			"Couldn't load runs",
		);
	});
});

describe('Runs page actions', () => {
	test('re-run is offered only where the server lists it, and a refusal is said', async () => {
		action = () => json({ code: 'forbidden', message: 'read-only' }, 403);
		await open();

		const rerun = [...document.querySelectorAll('button')].filter(
			(el) => el.textContent?.trim() === 'Re-run',
		);
		expect(rerun).toHaveLength(1);

		button('Re-run', card('main-deploy')).click();
		await settle();
		expect(
			card('main-deploy').querySelector('[role="alert"]')?.textContent,
		).toContain('Actions write');

		action = () => json({}, 202);
		button('Re-run', card('main-deploy')).click();
		await settle();

		expect(
			card('main-deploy').querySelector('[role="status"]')?.textContent,
		).toContain('Re-run requested');
		expect(posted).toEqual([
			'/api/runs/run-failed/rerun',
			'/api/runs/run-failed/rerun',
		]);
	});

	test('cancel is sent only after confirmation, and a refusal is said', async () => {
		action = () => json({ code: 'forbidden', message: 'read-only' }, 403);
		await open();

		const cancel = [...document.querySelectorAll('button')].filter(
			(el) => el.textContent?.trim() === 'Cancel run',
		);
		expect(cancel).toHaveLength(1);

		// Backing out sends nothing.
		button('Cancel run', card('preview-build')).click();
		await settle();
		expect(document.querySelector('[role="alertdialog"]')).not.toBeNull();
		button('Keep running', document.body).click();
		await settle();
		expect(posted).toEqual([]);

		button('Cancel run', card('preview-build')).click();
		await settle();
		const dialog = document.querySelector<HTMLElement>('[role="alertdialog"]');
		button('Cancel run', dialog ?? document).click();
		await settle();

		expect(
			card('preview-build').querySelector('[role="alert"]')?.textContent,
		).toContain('Actions write');
		expect(posted).toEqual(['/api/runs/run-running/cancel']);
	});
});

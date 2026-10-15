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

// The Flaky page against a fake network: what its window toggle asks the server
// for, and what it says when nothing is flaky, the list cannot load or a
// quarantine cannot be saved. The ranking and the rates are the server's; this is
// the page's own wiring.

const flaky = {
	window: '7d',
	steps: [
		{
			pipelineId: 'p-api',
			pipelineName: 'api-gateway',
			repoId: 'r1',
			name: 'Run unit tests',
			flakeRate: 0.25,
			runCount: 120,
			recentOutcomes: ['passed', 'failed', 'passed', 'passed'],
			quarantined: false,
		},
	],
	hasMore: false,
};

let network: ReturnType<typeof fakeNetwork>;
let respond: (url: URL) => Response;
let quarantine: (url: URL) => Response;
let app: ReturnType<typeof mount> | undefined;

function flakyRequests(): URL[] {
	return network.requested.filter((url) => url.pathname === '/api/steps/flaky');
}

beforeEach(() => {
	localStorage.clear();
	initTelemetryWindow();
	respond = () => json(flaky);
	quarantine = () => json({}, 500);

	network = fakeNetwork((url) => {
		if (url.pathname === '/api/steps/flaky') return respond(url);
		if (url.pathname === '/api/branches') return json({ branches: [] });
		if (url.pathname.endsWith('/quarantine')) return quarantine(url);
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

describe('Flaky page', () => {
	test('lists each flaky step it is given', async () => {
		await open();

		expect(document.body.textContent).toContain('Run unit tests');
		expect(document.body.textContent).toContain('api-gateway');
	});

	test('the window toggle asks the server for that window', async () => {
		await open();

		radio('30 days').click();
		await settle();

		expect(flakyRequests().at(-1)?.searchParams.get('window')).toBe('30d');
		expect(radio('30 days').getAttribute('aria-checked')).toBe('true');
	});

	test('with no saved window it asks for none, and shows the one the server used', async () => {
		respond = () => json({ ...flaky, window: '24h' });

		await open();

		expect(flakyRequests()[0]?.searchParams.has('window')).toBe(false);
		expect(radio('24 hours').getAttribute('aria-checked')).toBe('true');
	});

	test('says so when nothing is flaky', async () => {
		respond = () => json({ ...flaky, steps: [] });

		await open();

		expect(document.body.textContent).toContain(
			'No flaky steps in this window.',
		);
		expect(document.body.textContent).not.toContain('Run unit tests');
	});

	test('says so when the list cannot load', async () => {
		await open();

		respond = () => json({}, 500);
		radio('24 hours').click();
		await settle();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			"Couldn't load flaky steps",
		);
	});
});

function button(name: string, within: ParentNode = document): HTMLElement {
	const found = [...within.querySelectorAll<HTMLElement>('button')].find(
		(el) => el.textContent?.trim() === name,
	);
	if (!found) throw new Error(`no button named ${name}`);

	return found;
}

describe('Flaky page quarantine', () => {
	test('marks the step when the quarantine is saved', async () => {
		const day = 86_400_000;
		quarantine = () =>
			json({
				quarantined: true,
				quarantine: {
					note: '',
					quarantinedAt: new Date().toISOString(),
					expiresAt: new Date(Date.now() + 30 * day).toISOString(),
				},
			});

		await open();

		button('Quarantine').click();
		await settle();
		button('Quarantine this step').click();
		await settle();

		// The same word the failure test asserts is absent, so that assertion looks
		// at something that really does appear.
		expect(document.body.textContent).toContain('Quarantined');
		expect(document.querySelector('[role="alert"]')).toBeNull();
	});

	test('says so when the quarantine could not be saved, and marks nothing', async () => {
		await open();

		button('Quarantine').click();
		await settle();
		button('Quarantine this step').click();
		await settle();

		expect(document.querySelector('[role="alert"]')?.textContent).toContain(
			"Couldn't update",
		);
		expect(document.body.textContent).not.toContain('Quarantined');
		expect(
			network.requested.some((url) => url.pathname.endsWith('/quarantine')),
		).toBe(true);
	});
});
